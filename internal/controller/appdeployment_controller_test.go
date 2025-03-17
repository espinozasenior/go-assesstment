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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	deskreev1 "github.com/espinozasenior/go-assesstment.git/api/v1"
)

// TestFixture encapsulates the test environment for AppDeployment controller tests
type TestFixture struct {
	// Basic configuration
	Name           string
	Namespace      string
	Image          string
	MinReplicas    int32
	MaxReplicas    int32
	MemoryLimit    string
	NamespacedName types.NamespacedName

	// Test context
	Context    context.Context
	Reconciler *AppDeploymentReconciler

	// Test constants
	Timeout  time.Duration
	Interval time.Duration
}

// NewTestFixture creates a new test fixture with default values
func NewTestFixture() *TestFixture {
	name := "test-app"
	namespace := "default"
	return &TestFixture{
		Name:           name,
		Namespace:      namespace,
		Image:          "nginx:latest",
		MinReplicas:    1,
		MaxReplicas:    2,
		MemoryLimit:    "256Mi",
		Context:        context.Background(),
		NamespacedName: types.NamespacedName{Name: name, Namespace: namespace},
		Reconciler: &AppDeploymentReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		},
		Timeout:  time.Second * 10,
		Interval: time.Millisecond * 250,
	}
}

// CreateAppDeployment creates a test AppDeployment resource with the fixture's configuration
func (t *TestFixture) CreateAppDeployment() *deskreev1.AppDeployment {
	appDeployment := &deskreev1.AppDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      t.Name,
			Namespace: t.Namespace,
			Labels: map[string]string{
				"app": t.Name,
			},
		},
		Spec: deskreev1.AppDeploymentSpec{
			AppName:     t.Name,
			MemoryLimit: t.MemoryLimit,
			MinReplicas: t.MinReplicas,
			MaxReplicas: t.MaxReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": t.Name,
				},
			},
			Template: deskreev1.PodTemplateSpec{
				ObjectMeta: deskreev1.ObjectMeta{
					Labels: map[string]string{
						"app": t.Name,
					},
				},
				Spec: deskreev1.PodSpec{
					Containers: []deskreev1.Container{
						{
							Name:  "container-" + t.Name,
							Image: t.Image,
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

	Err := k8sClient.Create(t.Context, appDeployment)
	Expect(Err).NotTo(HaveOccurred())
	return appDeployment
}

// CleanupResources deletes all test resources created by the fixture
func (t *TestFixture) CleanupResources() {
	// Delete AppDeployment if it exists
	appDeployment := &deskreev1.AppDeployment{}
	err := k8sClient.Get(t.Context, t.NamespacedName, appDeployment)
	if err == nil {
		Expect(k8sClient.Delete(t.Context, appDeployment)).To(Succeed())
	}

	// Delete Deployment if it exists
	deployment := &appsv1.Deployment{}
	err = k8sClient.Get(t.Context, t.NamespacedName, deployment)
	if err == nil {
		Expect(k8sClient.Delete(t.Context, deployment)).To(Succeed())
	}
}

// WaitForResourceDeletion waits for the AppDeployment to be deleted
func (t *TestFixture) WaitForResourceDeletion() {
	Eventually(func() bool {
		err := k8sClient.Get(t.Context, t.NamespacedName, &deskreev1.AppDeployment{})
		return errors.IsNotFound(err)
	}, t.Timeout, t.Interval).Should(BeTrue())
}

// DeploymentExists checks if a deployment exists
func (t *TestFixture) DeploymentExists() bool {
	deployment := &appsv1.Deployment{}
	err := k8sClient.Get(t.Context, t.NamespacedName, deployment)
	return err == nil
}

// UpdateDeploymentStatus updates the status of a deployment
func (t *TestFixture) UpdateDeploymentStatus(availableReplicas, desiredReplicas int32) error {
	deployment := &appsv1.Deployment{}
	err := k8sClient.Get(t.Context, t.NamespacedName, deployment)
	if err != nil {
		return err
	}

	deployment.Status.AvailableReplicas = availableReplicas
	deployment.Status.Replicas = desiredReplicas
	deployment.Status.ReadyReplicas = availableReplicas
	return k8sClient.Status().Update(t.Context, deployment)
}

// GetAppDeploymentStatus gets the current status of the AppDeployment
func (t *TestFixture) GetAppDeploymentStatus() (string, error) {
	appDeployment := &deskreev1.AppDeployment{}
	err := k8sClient.Get(t.Context, t.NamespacedName, appDeployment)
	if err != nil {
		return "", err
	}
	return appDeployment.Status.State, nil
}

// ReconcileAppDeployment triggers a reconciliation
func (t *TestFixture) ReconcileAppDeployment() error {
	_, err := t.Reconciler.Reconcile(t.Context, reconcile.Request{
		NamespacedName: t.NamespacedName,
	})
	return err
}

// VerifyAppDeploymentStatus verifies the AppDeployment status matches the expected value
func (t *TestFixture) VerifyAppDeploymentStatus(expectedStatus string) {
	Eventually(func() string {
		status, err := t.GetAppDeploymentStatus()
		if err != nil {
			return ""
		}
		return status
	}, t.Timeout, t.Interval).Should(Equal(expectedStatus))
}

// SetupTest prepares the test environment by cleaning up existing resources
func (t *TestFixture) SetupTest() {
	t.CleanupResources()
	t.WaitForResourceDeletion()
}

var _ = Describe("AppDeployment Controller", func() {
	var fixture *TestFixture

	BeforeEach(func() {
		// Initialize test environment
		fixture = NewTestFixture()
		fixture.SetupTest()
	})

	AfterEach(func() {
		// Clean up resources after each test
		fixture.CleanupResources()
	})

	Context("When creating a new AppDeployment", func() {
		It("should create a deployment and update to Running state when replicas are available", func() {
			By("Creating a new AppDeployment resource")
			fixture.CreateAppDeployment()

			By("Reconciling the AppDeployment")
			Err := fixture.ReconcileAppDeployment()
			Expect(Err).NotTo(HaveOccurred())

			By("Verifying the Deployment was created")
			Eventually(func() bool {
				return fixture.DeploymentExists()
			}, fixture.Timeout, fixture.Interval).Should(BeTrue())

			By("Updating the Deployment status to have available replicas")
			Err = fixture.UpdateDeploymentStatus(1, 1)
			Expect(Err).NotTo(HaveOccurred())

			By("Reconciling the AppDeployment again")
			Err = fixture.ReconcileAppDeployment()
			Expect(Err).NotTo(HaveOccurred())

			By("Verifying the AppDeployment status was updated to Running")
			fixture.VerifyAppDeploymentStatus("Running")
		})

		It("should set status to Pending when deployment has zero replicas", func() {
			By("Creating a new AppDeployment resource")
			fixture.CreateAppDeployment()

			By("Reconciling the AppDeployment")
			Err := fixture.ReconcileAppDeployment()
			Expect(Err).NotTo(HaveOccurred())

			By("Verifying the Deployment was created")
			Eventually(func() bool {
				return fixture.DeploymentExists()
			}, fixture.Timeout, fixture.Interval).Should(BeTrue())

			By("Updating the Deployment status to have zero available replicas")
			Err = fixture.UpdateDeploymentStatus(0, 1)
			Expect(Err).NotTo(HaveOccurred())

			By("Reconciling the AppDeployment again")
			Err = fixture.ReconcileAppDeployment()
			Expect(Err).NotTo(HaveOccurred())

			By("Verifying the AppDeployment status was updated to Pending")
			fixture.VerifyAppDeploymentStatus("Pending")
		})

		It("should set status to Pending when deployment has fewer replicas than desired", func() {
			By("Creating a new AppDeployment resource with multiple replicas")
			fixture.MinReplicas = 2
			fixture.MaxReplicas = 4
			fixture.CreateAppDeployment()

			By("Reconciling the AppDeployment")
			Err := fixture.ReconcileAppDeployment()
			Expect(Err).NotTo(HaveOccurred())

			By("Verifying the Deployment was created")
			Eventually(func() bool {
				return fixture.DeploymentExists()
			}, fixture.Timeout, fixture.Interval).Should(BeTrue())

			By("Updating the Deployment status to have fewer replicas than desired")
			Err = fixture.UpdateDeploymentStatus(1, 2)
			Expect(Err).NotTo(HaveOccurred())

			By("Reconciling the AppDeployment again")
			Err = fixture.ReconcileAppDeployment()
			Expect(Err).NotTo(HaveOccurred())

			By("Verifying the AppDeployment status was updated to Pending")
			fixture.VerifyAppDeploymentStatus("Pending")
		})
	})
})
