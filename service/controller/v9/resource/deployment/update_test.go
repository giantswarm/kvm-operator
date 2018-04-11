package deployment

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/operatorkit/framework/context/updateallowedcontext"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/kvm-operator/service/kvmconfig/v9/key"
)

func Test_Resource_Deployment_newUpdateChange(t *testing.T) {
	testCases := []struct {
		Ctx                         context.Context
		Obj                         interface{}
		CurrentState                interface{}
		DesiredState                interface{}
		ExpectedDeploymentsToUpdate []*v1beta1.Deployment
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
			CurrentState:                []*v1beta1.Deployment{},
			DesiredState:                []*v1beta1.Deployment{},
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
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
			},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
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
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2-modified",
									},
								},
							},
						},
					},
				},
			},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
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

		// Test 3, is the same as 2 but with the version bundle version being
		// changed.
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
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.1.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
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

		// Test 4, is the same as 4 but with the version bundle version being
		// changed.
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
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
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

		// Test 5, when deployments should be updaated but their status is not "safe",
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
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
					Status: extensionsv1.DeploymentStatus{
						AvailableReplicas: 1,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
					Status: extensionsv1.DeploymentStatus{
						AvailableReplicas: 1,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
			},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
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
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
					Status: extensionsv1.DeploymentStatus{
						AvailableReplicas: 1,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
					Status: extensionsv1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
			},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
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
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
					Status: extensionsv1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
					Status: extensionsv1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
			},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
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
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
					Status: extensionsv1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.2.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
					Status: extensionsv1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
			},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
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
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
					Status: extensionsv1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
					Status: extensionsv1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
			},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
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

		// Test 10, is the same as 9 but with an empty version bundle version
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
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
					Status: extensionsv1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
					Status: extensionsv1.DeploymentStatus{
						AvailableReplicas: 2,
						ReadyReplicas:     2,
						Replicas:          2,
						UpdatedReplicas:   2,
					},
				},
			},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
						Annotations: map[string]string{
							key.VersionBundleVersionAnnotation: "1.3.0",
						},
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
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
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := DefaultConfig()
		resourceConfig.K8sClient = fake.NewSimpleClientset()
		resourceConfig.Logger = microloggertest.New()
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for _, tc := range testCases {
		updateState, err := newResource.newUpdateChange(tc.Ctx, tc.Obj, tc.CurrentState, tc.DesiredState)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		t.Run("deploymentsToUpdate", func(t *testing.T) {
			if tc.ExpectedDeploymentsToUpdate == nil {
				if updateState != nil {
					t.Fatalf("expected %#v got %#v", nil, updateState)
				}
			} else {
				deploymentsToUpdate, ok := updateState.([]*v1beta1.Deployment)
				if !ok {
					t.Fatalf("expected %T got %T", []*v1beta1.Deployment{}, updateState)
				}
				if !reflect.DeepEqual(deploymentsToUpdate, tc.ExpectedDeploymentsToUpdate) {
					t.Fatalf("expected %#v got %#v", tc.ExpectedDeploymentsToUpdate, deploymentsToUpdate)
				}
			}
		})
	}
}
