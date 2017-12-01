package service

import (
	"context"
	"testing"

	"github.com/giantswarm/clustertpr"
	clustertprspec "github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/micrologger/microloggertest"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func Test_Resource_Service_newCreateChange(t *testing.T) {
	testCases := []struct {
		Obj                  interface{}
		CurrentState         interface{}
		DesiredState         interface{}
		ExpectedServiceNames []string
	}{
		// Test 1, in case current state and desired state are empty the create
		// state should be empty.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState:         []*apiv1.Service{},
			DesiredState:         []*apiv1.Service{},
			ExpectedServiceNames: []string{},
		},

		// Test 2, in case current state equals desired state the create state
		// should be empty.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*apiv1.Service{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "service-1",
					},
				},
			},
			DesiredState: []*apiv1.Service{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "service-1",
					},
				},
			},
			ExpectedServiceNames: []string{},
		},

		// Test 3, in case current state misses one item of desired state the create
		// state should contain the missing item of the desired state.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*apiv1.Service{},
			DesiredState: []*apiv1.Service{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "service-1",
					},
				},
			},
			ExpectedServiceNames: []string{
				"service-1",
			},
		},

		// Test 4, in case current state misses items of desired state the create
		// state should contain the missing items of the desired state.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*apiv1.Service{},
			DesiredState: []*apiv1.Service{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "service-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "service-2",
					},
				},
			},
			ExpectedServiceNames: []string{
				"service-1",
				"service-2",
			},
		},

		// Test 5, in case current state contains one item not being in desired
		// state the create state should not contain the missing item of the desired
		// state.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*apiv1.Service{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "service-1",
					},
				},
			},
			DesiredState:         []*apiv1.Service{},
			ExpectedServiceNames: []string{},
		},

		// Test 6, in case current state contains items not being in desired state
		// the create state should not contain the missing items of the desired
		// state.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*apiv1.Service{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "service-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "service-2",
					},
				},
			},
			DesiredState:         []*apiv1.Service{},
			ExpectedServiceNames: []string{},
		},

		// Test 7, in case current state contains some items of desired state the
		// create state should contain the items being in desired state which are
		// not in create state.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*apiv1.Service{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "service-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "service-2",
					},
				},
			},
			DesiredState: []*apiv1.Service{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "service-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "service-2",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "service-3",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "service-4",
					},
				},
			},
			ExpectedServiceNames: []string{
				"service-3",
				"service-4",
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

	for i, tc := range testCases {
		result, err := newResource.newCreateChange(context.TODO(), tc.Obj, tc.CurrentState, tc.DesiredState)
		if err != nil {
			t.Fatalf("case %d expected %#v got %#v", i+1, nil, err)
		}

		configMaps, ok := result.([]*apiv1.Service)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*apiv1.Service{}, result)
		}

		if len(configMaps) != len(tc.ExpectedServiceNames) {
			t.Fatalf("case %d expected %d config maps got %d", i+1, len(tc.ExpectedServiceNames), len(configMaps))
		}
	}
}
