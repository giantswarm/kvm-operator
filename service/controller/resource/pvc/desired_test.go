package pvc

import (
	"context"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
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
			Obj: &v1alpha2.KVMCluster{
				Spec: v1alpha2.KVMClusterSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
						Masters: []v1alpha1.ClusterNode{
							{},
						},
						Workers: []v1alpha1.ClusterNode{
							{},
						},
					},
					KVM: v1alpha2.KVMClusterSpecKVM{
						K8sKVM: v1alpha2.KVMClusterSpecKVMK8sKVM{
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
			Obj: &v1alpha2.KVMCluster{
				Spec: v1alpha2.KVMClusterSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
						Masters: []v1alpha1.ClusterNode{
							{},
						},
						Workers: []v1alpha1.ClusterNode{
							{},
							{},
							{},
						},
					},
					KVM: v1alpha2.KVMClusterSpecKVM{
						K8sKVM: v1alpha2.KVMClusterSpecKVMK8sKVM{
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
			Obj: &v1alpha2.KVMCluster{
				Spec: v1alpha2.KVMClusterSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
						Masters: []v1alpha1.ClusterNode{
							{},
							{},
							{},
						},
						Workers: []v1alpha1.ClusterNode{
							{},
							{},
							{},
						},
					},
					KVM: v1alpha2.KVMClusterSpecKVM{
						K8sKVM: v1alpha2.KVMClusterSpecKVMK8sKVM{
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
			Obj: &v1alpha2.KVMCluster{
				Spec: v1alpha2.KVMClusterSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
						Masters: []v1alpha1.ClusterNode{
							{},
						},
						Workers: []v1alpha1.ClusterNode{
							{},
						},
					},
					KVM: v1alpha2.KVMClusterSpecKVM{
						K8sKVM: v1alpha2.KVMClusterSpecKVMK8sKVM{
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
			Obj: &v1alpha2.KVMCluster{
				Spec: v1alpha2.KVMClusterSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
						Masters: []v1alpha1.ClusterNode{
							{},
						},
						Workers: []v1alpha1.ClusterNode{
							{},
							{},
							{},
						},
					},
					KVM: v1alpha2.KVMClusterSpecKVM{
						K8sKVM: v1alpha2.KVMClusterSpecKVMK8sKVM{
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
			Obj: &v1alpha2.KVMCluster{
				Spec: v1alpha2.KVMClusterSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
						Masters: []v1alpha1.ClusterNode{
							{},
							{},
							{},
						},
						Workers: []v1alpha1.ClusterNode{
							{},
							{},
							{},
						},
					},
					KVM: v1alpha2.KVMClusterSpecKVM{
						K8sKVM: v1alpha2.KVMClusterSpecKVMK8sKVM{
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

		PVCs, ok := result.([]*corev1.PersistentVolumeClaim)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*corev1.PersistentVolumeClaim{}, result)
		}

		if testGetEtcdCount(PVCs) != tc.ExpectedEtcdCount {
			t.Fatalf("case %d expected %d worker nodes got %d", i+1, tc.ExpectedEtcdCount, testGetEtcdCount(PVCs))
		}
	}
}

func testGetEtcdCount(PVCs []*corev1.PersistentVolumeClaim) int {
	var count int

	for _, i := range PVCs {
		if strings.HasPrefix(i.Name, "pvc-master-etcd") {
			count++
		}
	}

	return count
}
