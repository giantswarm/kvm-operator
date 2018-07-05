package ingress

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	"k8s.io/api/extensions/v1beta1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_Ingress_newDeleteChange(t *testing.T) {
	testCases := []struct {
		Obj                  interface{}
		CurrentState         interface{}
		DesiredState         interface{}
		ExpectedIngressNames []string
	}{
		// Test 1, in case current state and desired state are empty the delete
		// state should be empty.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState:         []*v1beta1.Ingress{},
			DesiredState:         []*v1beta1.Ingress{},
			ExpectedIngressNames: []string{},
		},

		// Test 2, in case current state has one item and equals desired state the
		// delete state should equal the current state.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
			},
			DesiredState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
			},
			ExpectedIngressNames: []string{
				"ingress-1",
			},
		},

		// Test 3, in case current state misses one item of desired state the delete
		// state should not contain the missing item of the desired state.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*v1beta1.Ingress{},
			DesiredState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
			},
			ExpectedIngressNames: []string{},
		},

		// Test 4, in case current state misses items of desired state the delete
		// state should not contain the missing items of the desired state.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*v1beta1.Ingress{},
			DesiredState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-2",
					},
				},
			},
			ExpectedIngressNames: []string{},
		},

		// Test 5, in case current state contains one item and desired state is
		// empty the delete state should be empty.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
			},
			DesiredState:         []*v1beta1.Ingress{},
			ExpectedIngressNames: []string{},
		},

		// Test 6, in case current state contains items and desired state is empty
		// the delete state should be empty.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-2",
					},
				},
			},
			DesiredState:         []*v1beta1.Ingress{},
			ExpectedIngressNames: []string{},
		},

		// Test 7, in case all items of current state are in desired state and
		// desired state contains more items not being in current state the create
		// state should contain all items being in current state.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-2",
					},
				},
			},
			DesiredState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-2",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-3",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-4",
					},
				},
			},
			ExpectedIngressNames: []string{
				"ingress-1",
				"ingress-2",
			},
		},

		// Test 8, in case all items of desired state are in current state and
		// current state contains more items not being in desired state the create
		// state should contain all items being in desired state.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-2",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-3",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-4",
					},
				},
			},
			DesiredState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-2",
					},
				},
			},
			ExpectedIngressNames: []string{
				"ingress-1",
				"ingress-2",
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
		result, err := newResource.newDeleteChange(context.TODO(), tc.Obj, tc.CurrentState, tc.DesiredState)
		if err != nil {
			t.Fatalf("case %d expected %#v got %#v", i+1, nil, err)
		}

		configMaps, ok := result.([]*v1beta1.Ingress)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*v1beta1.Ingress{}, result)
		}

		if len(configMaps) != len(tc.ExpectedIngressNames) {
			t.Fatalf("case %d expected %d config maps got %d", i+1, len(tc.ExpectedIngressNames), len(configMaps))
		}
	}
}
