package deploymentv1

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/clustertpr"
	clustertprspec "github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/operatorkit/framework/context/updateallowedcontext"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	extensionsv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/giantswarm/kvm-operator/service/messagecontext"
)

func Test_Resource_Deployment_newUpdateChange(t *testing.T) {
	testCases := []struct {
		Ctx                         context.Context
		Obj                         interface{}
		CurrentState                interface{}
		DesiredState                interface{}
		ExpectedDeploymentsToUpdate []*v1beta1.Deployment
	}{
		// Test 1, in case current state and desired state are empty the update
		// state should be empty.
		{
			Ctx: context.TODO(),
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState:                []*v1beta1.Deployment{},
			DesiredState:                []*v1beta1.Deployment{},
			ExpectedDeploymentsToUpdate: nil,
		},

		// Test 2, in case current state and desired state are equal the update
		// state should be empty.
		{
			Ctx: context.TODO(),
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
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
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: nil,
		},

		// Test 3, in case current state contains two items and desired state
		// contains the same state but one object is modified internally the update
		// state should be empty in case updates are not allowed.
		{
			Ctx: context.TODO(),
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
									},
								},
							},
						},
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
										Name: "deployment-2-container-2-modified",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-2-config-map-2",
												},
											},
										},
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
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
									},
								},
							},
						},
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
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-2-config-map-2",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: nil,
		},

		// Test 4, in case current state contains two items and desired state
		// contains the same state but one object is modified internally the update
		// state should contain the the modified item from the current state.
		{
			Ctx: func() context.Context {
				ctx := context.Background()

				{
					m := messagecontext.NewMessage()
					m.ConfigMapNames = append(m.ConfigMapNames, "deployment-2-config-map-2")
					ctx = messagecontext.NewContext(ctx, m)
				}

				{
					ctx = updateallowedcontext.NewContext(ctx, make(chan struct{}))
					updateallowedcontext.SetUpdateAllowed(ctx)
				}

				return ctx
			}(),
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
									},
								},
							},
						},
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
										Name: "deployment-2-container-2-modified",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-2-config-map-2",
												},
											},
										},
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
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
									},
								},
							},
						},
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
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-2-config-map-2",
												},
											},
										},
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
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-2-config-map-2",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},

		// Test 5, same as 4 but ensuring the right deployments are computed as
		// update state when correspondig config names have changed.
		{
			Ctx: func() context.Context {
				ctx := context.Background()

				{
					m := messagecontext.NewMessage()
					m.ConfigMapNames = append(m.ConfigMapNames, "deployment-2-config-map-2")
					ctx = messagecontext.NewContext(ctx, m)
				}

				{
					ctx = updateallowedcontext.NewContext(ctx, make(chan struct{}))
					updateallowedcontext.SetUpdateAllowed(ctx)
				}

				return ctx
			}(),
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
									},
								},
							},
						},
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
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
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
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
									},
								},
							},
						},
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
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-2-config-map-2",
												},
											},
										},
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
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-2-config-map-2",
												},
											},
										},
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
			deploymentsToUpdate, ok := updateState.([]*v1beta1.Deployment)
			if !ok {
				t.Fatalf("expected %T got %T", []*v1beta1.Deployment{}, updateState)
			}
			if !reflect.DeepEqual(deploymentsToUpdate, tc.ExpectedDeploymentsToUpdate) {
				t.Fatalf("expected %#v got %#v", tc.ExpectedDeploymentsToUpdate, deploymentsToUpdate)
			}
		})
	}
}
