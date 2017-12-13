package ingressv2

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_Ingress_GetDesiredState(t *testing.T) {
	testCases := []struct {
		Obj               interface{}
		ExpectedAPICount  int
		ExpectedEtcdCount int
	}{
		// Test 1 ensures there is one ingress for master and worker each when there
		// is one master and one worker node in the custom object.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
						Masters: []v1alpha1.ClusterNode{
							{},
						},
						Workers: []v1alpha1.ClusterNode{
							{},
						},
					},
				},
			},
			ExpectedAPICount:  1,
			ExpectedEtcdCount: 1,
		},

		// Test 2 ensures there is one ingress for master and worker each when there
		// is one master and three worker nodes in the custom object.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
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
				},
			},
			ExpectedAPICount:  1,
			ExpectedEtcdCount: 1,
		},

		// Test 3 ensures there is one ingress for master and worker each when there
		// are three master and three worker nodes in the custom object.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
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
				},
			},
			ExpectedAPICount:  1,
			ExpectedEtcdCount: 1,
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

		ingresses, ok := result.([]*v1beta1.Ingress)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*v1beta1.Ingress{}, result)
		}

		if testGetAPICount(ingresses) != tc.ExpectedAPICount {
			t.Fatalf("case %d expected %d master nodes got %d", i+1, tc.ExpectedAPICount, testGetAPICount(ingresses))
		}

		if testGetEtcdCount(ingresses) != tc.ExpectedEtcdCount {
			t.Fatalf("case %d expected %d worker nodes got %d", i+1, tc.ExpectedEtcdCount, testGetEtcdCount(ingresses))
		}

		if len(ingresses) != tc.ExpectedAPICount+tc.ExpectedEtcdCount {
			t.Fatalf("case %d expected %d nodes got %d", i+1, tc.ExpectedAPICount+tc.ExpectedEtcdCount, len(ingresses))
		}
	}
}

func testGetAPICount(ingresses []*v1beta1.Ingress) int {
	var count int

	for _, i := range ingresses {
		if i.Name == "api" {
			count++
		}
	}

	return count
}

func testGetEtcdCount(ingresses []*v1beta1.Ingress) int {
	var count int

	for _, i := range ingresses {
		if i.Name == "etcd" {
			count++
		}
	}

	return count
}
