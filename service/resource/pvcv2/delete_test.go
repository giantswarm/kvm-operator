package pvcv2

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func Test_Resource_PVC_newDeleteChange(t *testing.T) {
	testCases := []struct {
		Obj              interface{}
		CurrentState     interface{}
		DesiredState     interface{}
		ExpectedPVCNames []string
	}{
		// Test 1, in case current state and desired state are empty the delete
		// state should be empty.
		{
			Obj: &v1alpha1.KvmConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState:     []*apiv1.PersistentVolumeClaim{},
			DesiredState:     []*apiv1.PersistentVolumeClaim{},
			ExpectedPVCNames: []string{},
		},

		// Test 2, in case current state has one item and equals desired state the
		// delete state should equal the current state.
		{
			Obj: &v1alpha1.KvmConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
			},
			DesiredState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
			},
			ExpectedPVCNames: []string{
				"pvc-1",
			},
		},

		// Test 3, in case current state misses one item of desired state the delete
		// state should not contain the missing item of the desired state.
		{
			Obj: &v1alpha1.KvmConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*apiv1.PersistentVolumeClaim{},
			DesiredState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
			},
			ExpectedPVCNames: []string{},
		},

		// Test 4, in case current state misses items of desired state the delete
		// state should not contain the missing items of the desired state.
		{
			Obj: &v1alpha1.KvmConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*apiv1.PersistentVolumeClaim{},
			DesiredState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-2",
					},
				},
			},
			ExpectedPVCNames: []string{},
		},

		// Test 5, in case current state contains one item and desired state is
		// empty the delete state should be empty.
		{
			Obj: &v1alpha1.KvmConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
			},
			DesiredState:     []*apiv1.PersistentVolumeClaim{},
			ExpectedPVCNames: []string{},
		},

		// Test 6, in case current state contains items and desired state is empty
		// the delete state should be empty.
		{
			Obj: &v1alpha1.KvmConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-2",
					},
				},
			},
			DesiredState:     []*apiv1.PersistentVolumeClaim{},
			ExpectedPVCNames: []string{},
		},

		// Test 7, in case all items of current state are in desired state and
		// desired state contains more items not being in current state the create
		// state should contain all items being in current state.
		{
			Obj: &v1alpha1.KvmConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-2",
					},
				},
			},
			DesiredState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-2",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-3",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-4",
					},
				},
			},
			ExpectedPVCNames: []string{
				"pvc-1",
				"pvc-2",
			},
		},

		// Test 8, in case all items of desired state are in current state and
		// current state contains more items not being in desired state the create
		// state should contain all items being in desired state.
		{
			Obj: &v1alpha1.KvmConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			CurrentState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-2",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-3",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-4",
					},
				},
			},
			DesiredState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-2",
					},
				},
			},
			ExpectedPVCNames: []string{
				"pvc-1",
				"pvc-2",
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

		configMaps, ok := result.([]*apiv1.PersistentVolumeClaim)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*apiv1.PersistentVolumeClaim{}, result)
		}

		if len(configMaps) != len(tc.ExpectedPVCNames) {
			t.Fatalf("case %d expected %d config maps got %d", i+1, len(tc.ExpectedPVCNames), len(configMaps))
		}
	}
}
