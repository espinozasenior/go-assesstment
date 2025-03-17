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

package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	deskreev1 "github.com/espinozasenior/go-assesstment.git/api/v1"
)

const (
	// StatePending indicates the deployment is in progress or waiting for resources
	StatePending = "Pending"
	// StateRunning indicates the deployment is active and running
	StateRunning = "Running"
	// StateFailed indicates the deployment has failed
	StateFailed = "Failed"
)

// AppDeploymentReconciler reconciles a AppDeployment object
type AppDeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=deskree.platform.deskree.com,resources=appdeployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=deskree.platform.deskree.com,resources=appdeployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=deskree.platform.deskree.com,resources=appdeployments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// It compares the state specified by the AppDeployment object against the actual cluster state,
// and then performs operations to make the cluster state reflect the state specified by the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *AppDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the AppDeployment instance
	appDeployment := &deskreev1.AppDeployment{}
	err := r.Get(ctx, req.NamespacedName, appDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			logger.Info("AppDeployment resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get AppDeployment")
		return ctrl.Result{}, err
	}

	// Check if the deployment exists
	deployment := &appsv1.Deployment{}
	deploymentName := appDeployment.Spec.AppName
	if deploymentName == "" {
		deploymentName = appDeployment.Name
	}

	err = r.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: req.Namespace}, deployment)

	// Update the AppDeployment status based on the deployment status
	if err != nil {
		if errors.IsNotFound(err) {
			// Deployment doesn't exist yet, create it
			logger.Info("Creating a new Deployment", "DeploymentName", deploymentName)

			// Create a new deployment based on the AppDeployment spec
			_, err = r.createDeploymentFromAppDeployment(appDeployment, req.Namespace)
			if err != nil {
				appDeployment.Status.State = StateFailed
				appDeployment.Status.Message = fmt.Sprintf("Failed to create deployment: %v", err)
				appDeployment.Status.AvailableReplicas = 0
				logger.Error(err, "Failed to create Deployment for AppDeployment", "DeploymentName", deploymentName)
				return ctrl.Result{}, err
			}

			// Set the deployment status
			appDeployment.Status.State = StatePending
			appDeployment.Status.Message = "Deployment created, waiting for replicas"
			appDeployment.Status.AvailableReplicas = 0
			logger.Info("Deployment created", "DeploymentName", deploymentName)
		} else {
			// Error getting deployment
			appDeployment.Status.State = StateFailed
			appDeployment.Status.Message = fmt.Sprintf("Error getting deployment: %v", err)
			appDeployment.Status.AvailableReplicas = 0
			logger.Error(err, "Failed to get Deployment for AppDeployment", "DeploymentName", deploymentName)
		}
	} else {
		availableReplicas := deployment.Status.AvailableReplicas
		desiredReplicas := *deployment.Spec.Replicas

		appDeployment.Status.AvailableReplicas = availableReplicas

		if availableReplicas == 0 {
			appDeployment.Status.State = StatePending
			appDeployment.Status.Message = "Deployment has no available replicas"
			logger.Info("Deployment has no available replicas", "DeploymentName", deploymentName)
		} else if availableReplicas < desiredReplicas {
			appDeployment.Status.State = StatePending
			appDeployment.Status.Message = fmt.Sprintf("Deployment is scaling up: %d/%d replicas available", availableReplicas, desiredReplicas)
			logger.Info("Deployment is scaling up", "DeploymentName", deploymentName, "AvailableReplicas", availableReplicas, "DesiredReplicas", desiredReplicas)
		} else {
			appDeployment.Status.State = StateRunning
			appDeployment.Status.Message = fmt.Sprintf("Deployment is active with %d replica(s)", availableReplicas)
			logger.Info("Deployment is running", "DeploymentName", deploymentName, "AvailableReplicas", availableReplicas)
		}
	}

	// Update the AppDeployment status
	err = r.Status().Update(ctx, appDeployment)
	if err != nil {
		logger.Error(err, "Failed to update AppDeployment status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&deskreev1.AppDeployment{}).
		Named("appdeployment").
		Complete(r)
}

// createDeploymentFromAppDeployment creates a new Deployment from an AppDeployment resource
func (r *AppDeploymentReconciler) createDeploymentFromAppDeployment(app *deskreev1.AppDeployment, namespace string) (*appsv1.Deployment, error) {
	// Set the deployment name
	deploymentName := app.Spec.AppName
	if deploymentName == "" {
		deploymentName = app.Name
	}

	// Set the replicas
	replicas := app.Spec.MinReplicas
	if replicas == 0 {
		replicas = 1 // Default to 1 replica if not specified
	}

	// Create the deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
			Labels:    app.Labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(app, deskreev1.GroupVersion.WithKind("AppDeployment")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: app.Spec.Selector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: app.Spec.Selector.MatchLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  app.Spec.Template.Spec.Containers[0].Name,
							Image: app.Spec.Template.Spec.Containers[0].Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: app.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort,
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse(app.Spec.MemoryLimit),
								},
							},
						},
					},
				},
			},
		},
	}

	// Create the deployment in the cluster
	err := r.Create(context.Background(), deployment)
	return deployment, err
}
