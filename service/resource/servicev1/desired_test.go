package servicev1

import (
	"context"
	"testing"

	"github.com/giantswarm/clustertpr"
	clustertprspec "github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/micrologger/microloggertest"
	"k8s.io/client-go/kubernetes/fake"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func Test_Resource_Service_GetDesiredState(t *testing.T) {
	testCases := []struct {
		Obj                 interface{}
		ExpectedMasterCount int
		ExpectedWorkerCount int
	}{
		// Test 1 ensures there is one service for master and worker each when there
		// is one master and one worker node in the custom object.
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
				},
			},
			ExpectedMasterCount: 1,
			ExpectedWorkerCount: 1,
		},

		// Test 2 ensures there is one service for master and worker each when there
		// is one master and three worker nodes in the custom object.
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
				},
			},
			ExpectedMasterCount: 1,
			ExpectedWorkerCount: 1,
		},

		// Test 3 ensures there is one service for master and worker each when there
		// are three master and three worker nodes in the custom object.
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
				},
			},
			ExpectedMasterCount: 1,
			ExpectedWorkerCount: 1,
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

		services, ok := result.([]*apiv1.Service)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*apiv1.Service{}, result)
		}

		if testGetMasterCount(services) != tc.ExpectedMasterCount {
			t.Fatalf("case %d expected %d master nodes got %d", i+1, tc.ExpectedMasterCount, testGetMasterCount(services))
		}

		if testGetWorkerCount(services) != tc.ExpectedWorkerCount {
			t.Fatalf("case %d expected %d worker nodes got %d", i+1, tc.ExpectedWorkerCount, testGetWorkerCount(services))
		}

		if len(services) != tc.ExpectedMasterCount+tc.ExpectedWorkerCount {
			t.Fatalf("case %d expected %d nodes got %d", i+1, tc.ExpectedMasterCount+tc.ExpectedWorkerCount, len(services))
		}
	}
}

func testGetMasterCount(services []*apiv1.Service) int {
	var count int

	for _, s := range services {
		if s.Name == "master" {
			count++
		}
	}

	return count
}

func testGetWorkerCount(services []*apiv1.Service) int {
	var count int

	for _, s := range services {
		if s.Name == "worker" {
			count++
		}
	}

	return count
}
