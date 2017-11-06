package pvc

import (
	"context"
	"strings"
	"testing"

	"github.com/giantswarm/clustertpr"
	clustertprspec "github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/kvmtpr"
	kvmtprspec "github.com/giantswarm/kvmtpr/spec"
	kvmtprspeckvm "github.com/giantswarm/kvmtpr/spec/kvm"
	"github.com/giantswarm/micrologger/microloggertest"
	"k8s.io/client-go/kubernetes/fake"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func Test_Resource_PVC_GetDesiredState(t *testing.T) {
	testCases := []struct {
		Obj               interface{}
		ExpectedEtcdCount int
	}{
		// Test 1 ensures there is one PVC for each master when there is one master
		// and one worker node and storage type is 'persistentVolume' in the custom
		// object.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
						Masters: []clustertprspec.Node{
							{},
						},
						Workers: []clustertprspec.Node{
							{},
						},
					},
					KVM: kvmtprspec.KVM{
						K8sKVM: kvmtprspeckvm.K8sKVM{
							StorageType: "persistentVolume",
						},
					},
				},
			},
			ExpectedEtcdCount: 1,
		},

		// Test 2 ensures there is one PVC for each master when there is one master
		// and three worker nodes and storage type is 'persistentVolume' in the
		// custom object.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
						Masters: []clustertprspec.Node{
							{},
						},
						Workers: []clustertprspec.Node{
							{},
							{},
							{},
						},
					},
					KVM: kvmtprspec.KVM{
						K8sKVM: kvmtprspeckvm.K8sKVM{
							StorageType: "persistentVolume",
						},
					},
				},
			},
			ExpectedEtcdCount: 1,
		},

		// Test 3 ensures there is one PVC for each master when there are three
		// master and three worker nodes and storage type is 'persistentVolume' in
		// the custom object.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
						Masters: []clustertprspec.Node{
							{},
							{},
							{},
						},
						Workers: []clustertprspec.Node{
							{},
							{},
							{},
						},
					},
					KVM: kvmtprspec.KVM{
						K8sKVM: kvmtprspeckvm.K8sKVM{
							StorageType: "persistentVolume",
						},
					},
				},
			},
			ExpectedEtcdCount: 3,
		},

		// Test 4 ensures there is no PVC for each master when there is one master
		// and one worker node and storage type is 'hostPath' in the custom
		// object.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
						Masters: []clustertprspec.Node{
							{},
						},
						Workers: []clustertprspec.Node{
							{},
						},
					},
					KVM: kvmtprspec.KVM{
						K8sKVM: kvmtprspeckvm.K8sKVM{
							StorageType: "hostPath",
						},
					},
				},
			},
			ExpectedEtcdCount: 0,
		},

		// Test 5 ensures there is no PVC for each master when there is one master
		// and three worker nodes and storage type is 'hostPath' in the
		// custom object.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
						Masters: []clustertprspec.Node{
							{},
						},
						Workers: []clustertprspec.Node{
							{},
							{},
							{},
						},
					},
					KVM: kvmtprspec.KVM{
						K8sKVM: kvmtprspeckvm.K8sKVM{
							StorageType: "hostPath",
						},
					},
				},
			},
			ExpectedEtcdCount: 0,
		},

		// Test 6 ensures there is no PVC for each master when there are three
		// master and three worker nodes and storage type is 'hostPath' in
		// the custom object.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
						Masters: []clustertprspec.Node{
							{},
							{},
							{},
						},
						Workers: []clustertprspec.Node{
							{},
							{},
							{},
						},
					},
					KVM: kvmtprspec.KVM{
						K8sKVM: kvmtprspeckvm.K8sKVM{
							StorageType: "hostPath",
						},
					},
				},
			},
			ExpectedEtcdCount: 0,
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

	for i, tc := range testCases {
		result, err := newResource.GetDesiredState(context.TODO(), tc.Obj)
		if err != nil {
			t.Fatalf("case %d expected %#v got %#v", i+1, nil, err)
		}

		PVCs, ok := result.([]*apiv1.PersistentVolumeClaim)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*apiv1.PersistentVolumeClaim{}, result)
		}

		if testGetEtcdCount(PVCs) != tc.ExpectedEtcdCount {
			t.Fatalf("case %d expected %d worker nodes got %d", i+1, tc.ExpectedEtcdCount, testGetEtcdCount(PVCs))
		}
	}
}

func testGetEtcdCount(PVCs []*apiv1.PersistentVolumeClaim) int {
	var count int

	for _, i := range PVCs {
		if strings.HasPrefix(i.Name, "pvc-master-etcd") {
			count++
		}
	}

	return count
}
