package deployment

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	apiextfake "github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/operatorkit/controller/context/updateallowedcontext"
	"github.com/giantswarm/tenantcluster"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stesting "k8s.io/client-go/testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func Test_Resource_Deployment_updateDeployments(t *testing.T) {
	clientset := setupReleasesClientSet()

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
				ctx := context.Background()

				{
					ctx = updateallowedcontext.NewContext(ctx, make(chan struct{}))
					updateallowedcontext.SetUpdateAllowed(ctx)
				}

				return ctx
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
		// Test 4, deployment with changed release version is being updated when there are
		// key component changes
		{
			Ctx: func() context.Context {
				ctx := context.Background()

				{
					ctx = updateallowedcontext.NewContext(ctx, make(chan struct{}))
					updateallowedcontext.SetUpdateAllowed(ctx)
				}

				return ctx
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

		// Test 5, deployment with changed release version is not being updated when there are
		// no key component changes
		{
			Ctx: func() context.Context {
				ctx := context.Background()

				{
					ctx = updateallowedcontext.NewContext(ctx, make(chan struct{}))
					updateallowedcontext.SetUpdateAllowed(ctx)
				}

				return ctx
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
							key.ReleaseVersionAnnotation:       "15.0.0",
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
		// Test 6, when deployments should be updated but their status is not "safe",
		// the update state should be empty.
		{
			Ctx: func() context.Context {
				ctx := context.Background()

				{
					ctx = updateallowedcontext.NewContext(ctx, make(chan struct{}))
					updateallowedcontext.SetUpdateAllowed(ctx)
				}

				return ctx
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

		// Test 7, is the same as 6 but with only one deployment not being "safe".
		{
			Ctx: func() context.Context {
				ctx := context.Background()

				{
					ctx = updateallowedcontext.NewContext(ctx, make(chan struct{}))
					updateallowedcontext.SetUpdateAllowed(ctx)
				}

				return ctx
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
		// Test 8, when all deployments are "safe" the update state should only
		// contain one deployment even though if multiple deployments should be
		// updated.
		{
			Ctx: func() context.Context {
				ctx := context.Background()

				{
					ctx = updateallowedcontext.NewContext(ctx, make(chan struct{}))
					updateallowedcontext.SetUpdateAllowed(ctx)
				}

				return ctx
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

		// Test 9, is based of 8 where the next deployment is ready to be updated.
		{
			Ctx: func() context.Context {
				ctx := context.Background()

				{
					ctx = updateallowedcontext.NewContext(ctx, make(chan struct{}))
					updateallowedcontext.SetUpdateAllowed(ctx)
				}

				return ctx
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

		// Test 10, is the same as 9 but ensures the update behaviour is preserved
		// even if no version bundle version annotation is present in the current
		// state.
		{
			Ctx: func() context.Context {
				ctx := context.Background()

				{
					ctx = updateallowedcontext.NewContext(ctx, make(chan struct{}))
					updateallowedcontext.SetUpdateAllowed(ctx)
				}

				return ctx
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
		// Test 11, is the same as 10 but ensures the update behaviour is preserved
		// even if no release version annotation is present in the current
		// state.
		{
			Ctx: func() context.Context {
				ctx := context.Background()

				{
					ctx = updateallowedcontext.NewContext(ctx, make(chan struct{}))
					updateallowedcontext.SetUpdateAllowed(ctx)
				}

				return ctx
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

		// Test 12, is the same as 11 but with an empty version bundle version
		// annotation.
		{
			Ctx: func() context.Context {
				ctx := context.Background()

				{
					ctx = updateallowedcontext.NewContext(ctx, make(chan struct{}))
					updateallowedcontext.SetUpdateAllowed(ctx)
				}

				return ctx
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
		// Test 13, is the same as 12 but with an empty release version
		// annotation.
		{
			Ctx: func() context.Context {
				ctx := context.Background()

				{
					ctx = updateallowedcontext.NewContext(ctx, make(chan struct{}))
					updateallowedcontext.SetUpdateAllowed(ctx)
				}

				return ctx
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

		// Test 14: if update is allowed but a tenant cluster master is unschedulable, do not update the worker deployment
		{
			Ctx: func() context.Context {
				ctx := context.Background()

				{
					ctx = updateallowedcontext.NewContext(ctx, make(chan struct{}))
					updateallowedcontext.SetUpdateAllowed(ctx)
				}

				return ctx
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

func getReleasesReactor(releases map[string]*releasev1alpha1.Release) k8stesting.Reactor {
	return &k8stesting.SimpleReactor{
		Verb:     "get",
		Resource: "releases",
		Reaction: func(action k8stesting.Action) (bool, runtime.Object, error) {
			getAction, ok := action.(k8stesting.GetActionImpl)
			if !ok {
				return false, nil, microerror.Maskf(wrongTypeError, "action != k8stesting.GetActionImpl")
			}

			releaseName := getAction.GetName()

			release, exists := releases[releaseName]
			if !exists {
				return false, nil, microerror.Mask(errors.New("release does not exist"))
			}
			return true, release, nil
		},
	}
}

func setupReleasesClientSet() *apiextfake.Clientset {
	release := releasev1alpha1.NewReleaseCR()
	release.ObjectMeta.Name = "v13.0.0"
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
			Name:    "cluster-operator",
			Version: "1.2.3",
		},
	}
	triggerUpdateRelease := releasev1alpha1.NewReleaseCR()
	triggerUpdateRelease.ObjectMeta.Name = "v14.0.0"
	triggerUpdateRelease.Spec.Components = []releasev1alpha1.ReleaseSpecComponent{
		{
			Name:    "kubernetes",
			Version: "1.15.13",
		},
		{
			Name:    "calico",
			Version: "3.9.1",
		},
		{
			Name:    "cluster-operator",
			Version: "1.2.3",
		},
	}
	dontTriggerUpdateRelease := releasev1alpha1.NewReleaseCR()
	dontTriggerUpdateRelease.ObjectMeta.Name = "v15.0.0"
	dontTriggerUpdateRelease.Spec.Components = []releasev1alpha1.ReleaseSpecComponent{
		{
			Name:    "kubernetes",
			Version: "1.15.11",
		},
		{
			Name:    "calico",
			Version: "3.9.1",
		},
		{
			Name:    "cluster-operator",
			Version: "1.2.5",
		},
	}

	releases := map[string]*releasev1alpha1.Release{
		"13.0.0": release,
		"14.0.0": triggerUpdateRelease,
		"15.0.0": dontTriggerUpdateRelease,
	}
	clientset := apiextfake.NewSimpleClientset()
	clientset.ReactionChain = append([]k8stesting.Reactor{
		getReleasesReactor(releases),
	}, clientset.ReactionChain...)

	return clientset
}
