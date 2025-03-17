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

// ObjectMeta is metadata that all persisted resources must have.
type ObjectMeta struct {
	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	Labels map[string]string `json:"labels,omitempty"`
}

// PodTemplateSpec describes the data a pod should have when created from a template
type PodTemplateSpec struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	ObjectMeta ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the pod.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Spec PodSpec `json:"spec,omitempty"`
}

// PodSpec is a description of a pod
type PodSpec struct {
	// List of containers belonging to the pod.
	// Containers cannot currently be added or removed.
	// There must be at least one container in a Pod.
	Containers []Container `json:"containers"`
}

// Container defines a single application container that is part of the pod.
type Container struct {
	// Name of the container specified as a DNS_LABEL.
	// Each container in a pod must have a unique name (DNS_LABEL).
	Name string `json:"name"`

	// Docker image name.
	Image string `json:"image"`

	// List of ports to expose from the container.
	Ports []ContainerPort `json:"ports,omitempty"`
}

// ContainerPort represents a network port in a single container.
type ContainerPort struct {
	// Number of port to expose on the pod's IP address.
	ContainerPort int32 `json:"containerPort"`
}
