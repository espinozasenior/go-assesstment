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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Server struct {
	client client.Client
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

	return &Server{client: client}, nil
}

func (s *Server) Start(port int) error {
	http.HandleFunc("/deploy", s.handleDeploy)
	http.HandleFunc("/status/", s.handleStatus)
	http.HandleFunc("/", s.handleDelete)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Server starting on %s", addr)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) handleDeploy(w http.ResponseWriter, r *http.Request) {
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

	if err := s.client.Create(context.Background(), appDeployment); err != nil {
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

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
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

	appDeployment := &deskreev1.AppDeployment{}
	if err := s.client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: "default"}, appDeployment); err != nil {
		http.Error(w, fmt.Sprintf("Failed to get AppDeployment: %v", err), http.StatusNotFound)
		return
	}

	response := StatusResponse{
		Status:   appDeployment.Status.State,
		Replicas: appDeployment.Status.AvailableReplicas,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
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

	if err := s.client.Delete(context.Background(), appDeployment); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete AppDeployment: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "AppDeployment \"%s\" deleted.\n", name)
}
