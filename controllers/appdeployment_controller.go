/*
Copyright 2023.

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

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	platformv1 "github.com/deskree-assessment/app-operator/api/v1"
)

// AppDeploymentReconciler reconciles a AppDeployment object
type AppDeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger zerolog.Logger
}

//+kubebuilder:rbac:groups=platform.deskree.com,resources=appdeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=platform.deskree.com,resources=appdeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=platform.deskree.com,resources=appdeployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *AppDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Logger.With().Str("appdeployment", req.NamespacedName.String()).Logger()
	logger.Info().Msg("Reconciling AppDeployment")

	// Fetch the AppDeployment instance
	appDeployment := &platformv1.AppDeployment{}
	err := r.Get(ctx, req.NamespacedName, appDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			logger.Info().Msg("AppDeployment resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error().Err(err).Msg("Failed to get AppDeployment")
		return ctrl.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: appDeployment.Name, Namespace: appDeployment.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForAppDeployment(appDeployment)
		logger.Info().Msg("Creating a new Deployment")
		err = r.Create(ctx, dep)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to create new Deployment")
			// Update status to Failed
			appDeployment.Status.Status = "Failed"
			appDeployment.Status.Message = fmt.Sprintf("Failed to create deployment: %s", err)
			err := r.updateStatus(ctx, req, appDeployment, logger)
			return ctrl.Result{}, err
		}
		// Update status to Pending
		appDeployment.Status.Status = "Pending"
		appDeployment.Status.Message = "Deployment created, waiting for pods"
		if err := r.updateStatus(ctx, req, appDeployment, logger); err != nil {
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error().Err(err).Msg("Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Ensure the deployment spec matches the desired state
	if deploymentNeedsUpdate(found, appDeployment) {
		found.Spec.Template.Spec.Containers[0].Image = appDeployment.Spec.Image
		found.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory] = resource.MustParse(appDeployment.Spec.MemoryLimit)
		found.Spec.Replicas = &appDeployment.Spec.MinReplicas

		logger.Info().Msg("Updating Deployment")
		err = r.Update(ctx, found)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to update Deployment")
			// Update status to Failed
			appDeployment.Status.Status = "Failed"
			appDeployment.Status.Message = fmt.Sprintf("Failed to update deployment: %s", err)
			err := r.updateStatus(ctx, req, appDeployment, logger)
			return ctrl.Result{}, err
		}
		// Update status to Pending
		appDeployment.Status.Status = "Pending"
		appDeployment.Status.Message = "Deployment updated, waiting for pods"
		if err := r.updateStatus(ctx, req, appDeployment, logger); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Determine the new status based on deployment state
	var newStatus string
	var newMessage string

	if found.Status.ReadyReplicas > 0 {
		newStatus = "Running"
		newMessage = fmt.Sprintf("Deployment is active with %d/%d replica(s) ready", found.Status.ReadyReplicas, found.Status.Replicas)
	} else if found.Status.UnavailableReplicas > 0 {
		newStatus = "Pending"
		newMessage = fmt.Sprintf("Deployment has %d unavailable replica(s)", found.Status.UnavailableReplicas)
	} else if found.Status.Replicas == 0 {
		newStatus = "Pending"
		newMessage = "Deployment created, waiting for pods to be scheduled"
	} else {
		newStatus = "Pending"
		newMessage = "Deployment created, waiting for pods"
	}

	// Only update status if it has changed
	if appDeployment.Status.Status != newStatus || appDeployment.Status.Message != newMessage {
		appDeployment.Status.Status = newStatus
		appDeployment.Status.Message = newMessage

		err = r.updateStatus(ctx, req, appDeployment, logger)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// deploymentForAppDeployment returns a deployment object for the specified AppDeployment
func (r *AppDeploymentReconciler) deploymentForAppDeployment(app *platformv1.AppDeployment) *appsv1.Deployment {
	labels := map[string]string{
		"app":        app.Spec.AppName,
		"controller": app.Name,
	}

	// Set the replica count to MinReplicas
	replicas := app.Spec.MinReplicas

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  app.Spec.AppName,
						Image: app.Spec.Image,
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
							Name:          "http",
						}},
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse(app.Spec.MemoryLimit),
							},
						},
					}},
				},
			},
		},
	}

	// Set AppDeployment instance as the owner and controller
	controllerutil.SetControllerReference(app, dep, r.Scheme)
	return dep
}

// deploymentNeedsUpdate checks if the deployment needs to be updated
func deploymentNeedsUpdate(dep *appsv1.Deployment, app *platformv1.AppDeployment) bool {
	if dep.Spec.Template.Spec.Containers[0].Image != app.Spec.Image {
		return true
	}

	if dep.Spec.Replicas != nil && *dep.Spec.Replicas != app.Spec.MinReplicas {
		return true
	}

	currentMemory := dep.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory]
	requiredMemory := resource.MustParse(app.Spec.MemoryLimit)

	if currentMemory.Cmp(requiredMemory) != 0 {
		return true
	}

	return false
}

// updateStatus updates the status of the AppDeployment with retry logic for conflict errors
func (r *AppDeploymentReconciler) updateStatus(ctx context.Context, req ctrl.Request, appDeployment *platformv1.AppDeployment, logger zerolog.Logger) error {
	// First, get the latest version of the resource to compare status
	latestAppDeployment := &platformv1.AppDeployment{}
	if err := r.Get(ctx, req.NamespacedName, latestAppDeployment); err != nil {
		logger.Error().Err(err).Msg("Failed to get latest AppDeployment for status comparison")
		return err
	}

	// Check if status has actually changed before updating
	if latestAppDeployment.Status.Status == appDeployment.Status.Status &&
		latestAppDeployment.Status.Message == appDeployment.Status.Message {
		logger.Debug().Msg("Status unchanged, skipping update")
		return nil
	}

	// Status has changed, proceed with update
	var updateErr error
	backoff := time.Second
	for retries := 0; retries < 5; retries++ {
		// Get the latest version again to avoid conflicts
		if retries > 0 {
			if err := r.Get(ctx, req.NamespacedName, latestAppDeployment); err != nil {
				logger.Error().Err(err).Msg("Failed to get latest AppDeployment")
				return err
			}
		}

		// Copy only status fields to avoid overwriting spec changes
		latestAppDeployment.Status.Status = appDeployment.Status.Status
		latestAppDeployment.Status.Message = appDeployment.Status.Message

		// Update status
		updateErr = r.Status().Update(ctx, latestAppDeployment)
		if updateErr == nil {
			return nil
		}

		if !errors.IsConflict(updateErr) {
			logger.Error().Err(updateErr).Msg("Failed to update AppDeployment status")
			return updateErr
		}

		// If we get a conflict, log and retry
		logger.Info().Msg("Conflict detected when updating AppDeployment status, retrying with latest version")

		// Wait before retrying
		time.Sleep(backoff)
		backoff *= 2 // Exponential backoff
	}

	logger.Error().Err(updateErr).Msg("Failed to update AppDeployment status after multiple retries")
	return updateErr
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1.AppDeployment{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
