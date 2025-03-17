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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AppDeploymentSpec defines the desired state of AppDeployment.
type AppDeploymentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Image is the container image to deploy
	Image string `json:"image,omitempty"`
	// AppName is the name of the application
	AppName string `json:"appName,omitempty"`
	// MemoryLimit specifies the memory limit for the container
	MemoryLimit string `json:"memoryLimit,omitempty"`
	// MinReplicas is the minimum number of replicas for the deployment
	MinReplicas int32 `json:"minReplicas,omitempty"`
	// MaxReplicas is the maximum number of replicas for the deployment
	MaxReplicas int32 `json:"maxReplicas,omitempty"`
	// Selector is the label selector for pods
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
	// Template is the pod template specification
	Template PodTemplateSpec `json:"template,omitempty"`
}

// AppDeploymentStatus defines the observed state of AppDeployment.
type AppDeploymentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// State represents the current state of the AppDeployment (Running, Pending, Failed)
	State string `json:"state,omitempty"`
	// Message provides additional information about the current state
	Message string `json:"message,omitempty"`
	// AvailableReplicas represents the number of replicas that are available
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`
	// Conditions represents the latest available observations of AppDeployment's current state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AppDeployment is the Schema for the appdeployments API.
type AppDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppDeploymentSpec   `json:"spec,omitempty"`
	Status AppDeploymentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AppDeploymentList contains a list of AppDeployment.
type AppDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppDeployment{}, &AppDeploymentList{})
}
