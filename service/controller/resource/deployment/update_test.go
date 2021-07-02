package deployment

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	apiextfake "github.com/giantswarm/apiextensions/v3/pkg/clientset/versioned/fake"
	"github.com/giantswarm/certs/v3/pkg/certs"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
)

func Test_Resource_Deployment_updateDeployments(t *testing.T) {
	// Create a fake release
	release := releasev1alpha1.NewReleaseCR()
	release.ObjectMeta.Name = "v1.0.3"
	release.Spec.Components = []releasev1alpha1.ReleaseSpecComponent{
		{
			Name:    "kubernetes",
			Version: "1.15.11",
		},
		{
			Name:    "calico",
			Version: "3.9.1",
		},
		{
			Name:    "etcd",
			Version: "3.3.15",
		},
	}
	clientset := apiextfake.NewSimpleClientset(release)

	testCases := []struct {
		Ctx                         context.Context
		Obj                         interface{}
		CurrentState                interface{}
		DesiredState                interface{}
		ExpectedDeploymentsToUpdate []*v1.Deployment
		FakeTCObjects               []runtime.Object
	}{
		// Test 0, in case current state and desired state are empty the update
		// state should be empty.
		{
			Ctx: context.TODO(),
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState:                []*v1.Deployment{},
			DesiredState:                []*v1.Deployment{},
			ExpectedDeploymentsToUpdate: nil,
		},

		// Test 1, in case current state and desired state are equal the update
		// state should be empty.
		{
			Ctx: context.TODO(),
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
			},
			DesiredState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: nil,
		},

		// Test 2, in case current state is modified, the update state should
		// be empty in case updates are not allowed.
		{
			Ctx: context.TODO(),
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2-modified",
									},
								},
							},
						},
					},
				},
			},
			DesiredState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: nil,
		},

		// Test 3, the deployment with changed bundle version is being updated.
		{
			Ctx: func() context.Context {
				return context.Background()
			}(),
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			DesiredState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
		},
		// Test 4, the deployment with changed release version is being updated.
		{
			Ctx: func() context.Context {
				return context.Background()
			}(),
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			DesiredState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "14.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "14.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
		},

		// Test 5, when deployments should be updated but their status is not "safe",
		// the update state should be empty.
		{
			Ctx: func() context.Context {
				return context.Background()
			}(),
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
					Status: v1.DeploymentStatus{
						AvailableReplicas: 1,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
					Status: v1.DeploymentStatus{
						AvailableReplicas: 1,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
			},
			DesiredState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: nil,
		},

		// Test 6, is the same as 5 but with only one deployment not being "safe".
		{
			Ctx: func() context.Context {
				return context.Background()
			}(),
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
					Status: v1.DeploymentStatus{
						AvailableReplicas: 1,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
					Status: v1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
			},
			DesiredState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: nil,
		},
		// Test 7, when all deployments are "safe" the update state should only
		// contain one deployment even though if multiple deployments should be
		// updated.
		{
			Ctx: func() context.Context {
				return context.Background()
			}(),
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
					Status: v1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
					Status: v1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
			},
			DesiredState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
			},
		},

		// Test 8, is based of 7 where the next deployment is ready to be updated.
		{
			Ctx: func() context.Context {
				return context.Background()
			}(),
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
					Status: v1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
					Status: v1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
			},
			DesiredState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
		},

		// Test 9, is the same as 8 but ensures the update behaviour is preserved
		// even if no version bundle version annotation is present in the current
		// state.
		{
			Ctx: func() context.Context {
				return context.Background()
			}(),
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
					Status: v1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation: "13.0.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
					Status: v1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
			},
			DesiredState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
		},
		// Test 10, is the same as 9 but ensures the update behaviour is preserved
		// even if no release version annotation is present in the current
		// state.
		{
			Ctx: func() context.Context {
				return context.Background()
			}(),
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
					Status: v1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
					Status: v1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
			},
			DesiredState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
		},

		// Test 11, is the same as 10 but with an empty version bundle version
		// annotation.
		{
			Ctx: func() context.Context {
				return context.Background()
			}(),
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
					Status: v1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
					Status: v1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
			},
			DesiredState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
		},
		// Test 12, is the same as 11 but with an empty release version
		// annotation.
		{
			Ctx: func() context.Context {
				return context.Background()
			}(),
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
					Status: v1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
					Status: v1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
			},
			DesiredState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
		},

		// Test 13: if update is allowed but a tenant cluster master is unschedulable, do not update the worker deployment
		{
			Ctx: func() context.Context {
				return context.Background()
			}(),
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
						Labels: map[string]string{"app": "master"},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
					Status: v1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
						Labels: map[string]string{"app": "worker"},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
					Status: v1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
			},
			DesiredState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
						Labels: map[string]string{"app": "master"},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.ReleaseVersionAnnotation:       "13.0.0",
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
						Labels: map[string]string{"app": "worker"},
					},
					Spec: v1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: nil,
			FakeTCObjects: []runtime.Object{
				// Create a master Node with a NoSchedule taint
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"role": "master"}},
					Spec: corev1.NodeSpec{
						Taints: []corev1.Taint{corev1.Taint{
							Key:    "NoSchedule",
							Effect: corev1.TaintEffectNoSchedule,
						}},
					},
				},
			},
		},
	}

	var err error
	logger := microloggertest.New()

	var certsSearcher certs.Interface
	{
		c := certs.Config{
			K8sClient:    fake.NewSimpleClientset(),
			Logger:       logger,
			WatchTimeout: 5 * time.Second,
		}

		certsSearcher, err = certs.NewSearcher(c)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	// This tenantcluster is not actually used. It exists to avoid errors when it is nil.
	// Each test case creates a new client for the tenant cluster and uses that for the test.
	var tenantCluster tenantcluster.Interface
	{
		c := tenantcluster.Config{
			CertsSearcher: certsSearcher,
			Logger:        logger,
			CertID:        certs.APICert,
		}

		tenantCluster, err = tenantcluster.New(c)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}
	var newResource *Resource
	{
		resourceConfig := Config{
			DNSServers:    "dnsserver1,dnsserver2",
			G8sClient:     clientset,
			K8sClient:     fake.NewSimpleClientset(),
			Logger:        microloggertest.New(),
			TenantCluster: tenantCluster,
		}
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for _, tc := range testCases {

		// This tenant client is actually used during the test.
		tenantK8sClient := fake.NewSimpleClientset(tc.FakeTCObjects...) // Pass in any fake TC objects

		updateState, err := newResource.updateDeployments(tc.Ctx, tc.CurrentState, tc.DesiredState, tenantK8sClient)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		t.Run("deploymentsToUpdate", func(t *testing.T) {
			if tc.ExpectedDeploymentsToUpdate == nil {
				if updateState != nil {
					t.Fatalf("expected %#v got %#v", nil, updateState)
				}
			} else {
				deploymentsToUpdate, ok := updateState.([]*v1.Deployment)
				if !ok {
					t.Fatalf("expected %T got %T", []*v1.Deployment{}, updateState)
				}
				if !reflect.DeepEqual(deploymentsToUpdate, tc.ExpectedDeploymentsToUpdate) {
					t.Fatalf("expected %#v got %#v", tc.ExpectedDeploymentsToUpdate, deploymentsToUpdate)
				}
			}
		})
	}
}
