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

package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "github.com/espinozasenior/go-assesstment.git/api/v1"
	"github.com/espinozasenior/go-assesstment.git/internal/apiserver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TestCacheHandlingAfterDelete tests that the cache is properly updated after a delete operation
func TestCacheHandlingAfterDelete(t *testing.T) {
	// Create a fake client with the AppDeployment scheme
	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed to add v1 scheme: %v", err)
	}

	// Create a test AppDeployment
	appName := "test-app"
	appDeployment := &v1.AppDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: "default",
		},
		Status: v1.AppDeploymentStatus{
			State:             "Running",
			AvailableReplicas: 1,
		},
	}

	// Create a fake client with the test AppDeployment
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(appDeployment).
		Build()

	// Create a server with the fake client
	server := &apiserver.Server{
		Client:          fakeClient,
		DeploymentCache: make(map[string]*v1.AppDeployment),
	}

	// Add the AppDeployment to the cache
	server.DeploymentCache[appName] = appDeployment

	// Test 1: Verify the AppDeployment is in the cache
	statusReq := httptest.NewRequest("GET", "/status/"+appName, nil)
	statusRecorder := httptest.NewRecorder()

	server.HandleStatus(statusRecorder, statusReq)

	if statusRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, statusRecorder.Code)
	}

	var statusResp apiserver.StatusResponse
	if err := json.NewDecoder(statusRecorder.Body).Decode(&statusResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if statusResp.Status != "Running" || statusResp.Replicas != 1 {
		t.Errorf("Expected status 'Running' with 1 replica, got status '%s' with %d replicas",
			statusResp.Status, statusResp.Replicas)
	}

	// Test 2: Delete the AppDeployment
	deleteReq := httptest.NewRequest("DELETE", "/"+appName, nil)
	deleteRecorder := httptest.NewRecorder()

	server.HandleDelete(deleteRecorder, deleteReq)

	if deleteRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, deleteRecorder.Code)
	}

	// Test 3: Verify the AppDeployment is removed from the cache
	// The API server should try to get it from the API and fail with NotFound
	statusReq2 := httptest.NewRequest("GET", "/status/"+appName, nil)
	statusRecorder2 := httptest.NewRecorder()

	server.HandleStatus(statusRecorder2, statusReq2)

	if statusRecorder2.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, statusRecorder2.Code)
	}

	// Verify the AppDeployment is not in the cache
	if _, exists := server.DeploymentCache[appName]; exists {
		t.Errorf("AppDeployment should have been removed from cache after deletion")
	}
}
