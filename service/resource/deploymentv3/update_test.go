package deploymentv3

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
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

		// Test 2, in case current state and desired state are equal the update
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
