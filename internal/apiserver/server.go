/*
Copyright 2025 LuisEspinoza.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	deskreev1 "github.com/espinozasenior/go-assesstment.git/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Server struct {
	Client  client.Client
	watcher watch.Interface
	stopCh  chan struct{}
	// Cache to store AppDeployment status information
	DeploymentCache map[string]*deskreev1.AppDeployment
}

type DeployRequest struct {
	Image       string `json:"image"`
	Name        string `json:"name"`
	MemoryLimit string `json:"memoryLimit"`
	MinReplicas int32  `json:"minReplicas"`
	MaxReplicas int32  `json:"maxReplicas"`
}

type StatusResponse struct {
	Status   string `json:"status"`
	Replicas int32  `json:"replicas"`
}

func NewServer() (*Server, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting kubeconfig: %v", err)
	}

	scheme := scheme.Scheme
	if err := deskreev1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("error adding deskree scheme: %v", err)
	}

	client, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("error creating client: %v", err)
	}

	server := &Server{
		Client:          client,
		stopCh:          make(chan struct{}),
		DeploymentCache: make(map[string]*deskreev1.AppDeployment),
	}

	// Setup the watcher for AppDeployment CRDs
	if err := server.setupWatcher(); err != nil {
		return nil, fmt.Errorf("error setting up watcher: %v", err)
	}

	return server, nil
}

func (s *Server) Start(port int) error {
	http.HandleFunc("/deploy", s.HandleDeploy)
	http.HandleFunc("/status/", s.HandleStatus)
	http.HandleFunc("/", s.HandleDelete)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Server starting on %s", addr)
	return http.ListenAndServe(addr, nil)
}

// setupWatcher creates a watch.Interface to monitor AppDeployment CRD changes
func (s *Server) setupWatcher() error {
	// Create a dynamic client to watch for AppDeployment resources
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("error getting kubeconfig: %v", err)
	}

	// Create a dynamic client
	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("error creating dynamic client: %v", err)
	}

	// Define the GVR (GroupVersionResource) for AppDeployment
	appDeploymentGVR := schema.GroupVersionResource{
		Group:    "deskree.platform.deskree.com",
		Version:  "v1",
		Resource: "appdeployments",
	}

	// Create a watcher for AppDeployment resources
	watcher, err := dynamicClient.Resource(appDeploymentGVR).Namespace("").Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error creating watcher: %v", err)
	}

	s.watcher = watcher

	// Start a goroutine to handle watch events
	go s.watchEvents()

	return nil
}

// watchEvents processes events from the watcher
func (s *Server) watchEvents() {
	log.Println("Starting to watch AppDeployment CRD changes")
	defer log.Println("Stopped watching AppDeployment CRD changes")

	for {
		select {
		case event, ok := <-s.watcher.ResultChan():
			if !ok {
				log.Println("Watcher channel closed, restarting watcher...")
				// Try to restart the watcher
				if err := s.setupWatcher(); err != nil {
					log.Printf("Failed to restart watcher: %v", err)
					return
				}
				continue
			}

			// Process the event based on its type
			switch event.Type {
			case watch.Added, watch.Modified:
				// Convert the unstructured object to AppDeployment
				unstructured, ok := event.Object.(*metav1.PartialObjectMetadata)
				if !ok {
					log.Printf("Unexpected object type: %T", event.Object)
					continue
				}

				// Get the AppDeployment from the API server to ensure we have the latest state
				appDeployment := &deskreev1.AppDeployment{}
				err := s.Client.Get(context.Background(), types.NamespacedName{Name: unstructured.GetName(), Namespace: unstructured.GetNamespace()}, appDeployment)
				if err != nil {
					log.Printf("Error getting AppDeployment %s/%s: %v", unstructured.GetNamespace(), unstructured.GetName(), err)
					continue
				}

				// Store the AppDeployment in the cache
				s.DeploymentCache[appDeployment.Name] = appDeployment
				log.Printf("AppDeployment %s updated in cache: state=%s, replicas=%d",
					appDeployment.Name, appDeployment.Status.State, appDeployment.Status.AvailableReplicas)

			case watch.Deleted:
				unstructured, ok := event.Object.(*metav1.PartialObjectMetadata)
				if !ok {
					log.Printf("Unexpected object type: %T", event.Object)
					continue
				}

				// Remove the AppDeployment from the cache
				delete(s.DeploymentCache, unstructured.GetName())
				log.Printf("AppDeployment %s removed from cache", unstructured.GetName())

			case watch.Error:
				log.Printf("Error event received: %v", event.Object)
			}
		case <-s.stopCh:
			return
		}
	}
}

func (s *Server) HandleDeploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Image == "" || req.MemoryLimit == "" {
		http.Error(w, "Name, image, and memoryLimit are required", http.StatusBadRequest)
		return
	}

	if req.MinReplicas <= 0 {
		req.MinReplicas = 1
	}

	if req.MaxReplicas < req.MinReplicas {
		req.MaxReplicas = req.MinReplicas
	}

	appDeployment := &deskreev1.AppDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: "default",
			Labels: map[string]string{
				"app.kubernetes.io/name":       req.Name,
				"app.kubernetes.io/managed-by": "go-assessment-api",
			},
		},
		Spec: deskreev1.AppDeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": req.Name,
				},
			},
			MemoryLimit: req.MemoryLimit,
			MinReplicas: req.MinReplicas,
			MaxReplicas: req.MaxReplicas,
			Template: deskreev1.PodTemplateSpec{
				ObjectMeta: deskreev1.ObjectMeta{
					Labels: map[string]string{
						"app": req.Name,
					},
				},
				Spec: deskreev1.PodSpec{
					Containers: []deskreev1.Container{
						{
							Name:  req.Name,
							Image: req.Image,
							Ports: []deskreev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	if err := s.Client.Create(context.Background(), appDeployment); err != nil {
		response := map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to create AppDeployment: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Deployment CRD %s created", req.Name),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	name := pathParts[2]
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	var appDeployment *deskreev1.AppDeployment
	var found bool

	// Check the cache first
	appDeployment, found = s.DeploymentCache[name]

	// If not found in cache, get it from the API server
	if !found {
		log.Printf("AppDeployment %s not found in cache, fetching from API server", name)
		appDeployment = &deskreev1.AppDeployment{}
		if err := s.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: "default"}, appDeployment); err != nil {
			http.Error(w, fmt.Sprintf("Failed to get AppDeployment: %v", err), http.StatusNotFound)
			return
		}
		// Add to cache for future requests
		s.DeploymentCache[name] = appDeployment
	} else {
		log.Printf("AppDeployment %s found in cache", name)
	}

	response := StatusResponse{
		Status:   appDeployment.Status.State,
		Replicas: appDeployment.Status.AvailableReplicas,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) HandleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	name := pathParts[1]
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	appDeployment := &deskreev1.AppDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
	}

	if err := s.Client.Delete(context.Background(), appDeployment); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete AppDeployment: %v", err), http.StatusInternalServerError)
		return
	}

	// Remove the deleted AppDeployment from the cache to prevent stale data
	if _, exists := s.DeploymentCache[name]; exists {
		delete(s.DeploymentCache, name)
		log.Printf("AppDeployment %s removed from cache after deletion", name)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "AppDeployment \"%s\" deleted.\n", name)
}
