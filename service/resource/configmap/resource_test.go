package configmap

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/giantswarm/certificatetpr/certificatetprtest"
	"github.com/giantswarm/clustertpr"
	clustertprspec "github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/operatorkit/framework/updateallowedcontext"
	"github.com/giantswarm/operatorkit/framework/updatenecessarycontext"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func Test_Resource_CloudConfig_GetDesiredState(t *testing.T) {
	testCases := []struct {
		Obj                 interface{}
		ExpectedMasterCount int
		ExpectedWorkerCount int
	}{
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
			ExpectedWorkerCount: 3,
		},
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
			ExpectedMasterCount: 3,
			ExpectedWorkerCount: 3,
		},
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := DefaultConfig()
		resourceConfig.CertWatcher = certificatetprtest.NewService()
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

		configMaps, ok := result.([]*apiv1.ConfigMap)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*apiv1.ConfigMap{}, result)
		}

		if testGetMasterCount(configMaps) != tc.ExpectedMasterCount {
			t.Fatalf("case %d expected %d master nodes got %d", i+1, tc.ExpectedMasterCount, testGetMasterCount(configMaps))
		}

		if testGetWorkerCount(configMaps) != tc.ExpectedWorkerCount {
			t.Fatalf("case %d expected %d worker nodes got %d", i+1, tc.ExpectedWorkerCount, testGetWorkerCount(configMaps))
		}

		if len(configMaps) != tc.ExpectedMasterCount+tc.ExpectedWorkerCount {
			t.Fatalf("case %d expected %d nodes got %d", i+1, tc.ExpectedMasterCount+tc.ExpectedWorkerCount, len(configMaps))
		}
	}
}

func Test_Resource_CloudConfig_GetCreateState(t *testing.T) {
	testCases := []struct {
		Obj                    interface{}
		CurrentState           interface{}
		DesiredState           interface{}
		ExpectedConfigMapNames []string
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
			CurrentState:           []*apiv1.ConfigMap{},
			DesiredState:           []*apiv1.ConfigMap{},
			ExpectedConfigMapNames: []string{},
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
			CurrentState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
				},
			},
			DesiredState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
				},
			},
			ExpectedConfigMapNames: []string{},
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
			CurrentState: []*apiv1.ConfigMap{},
			DesiredState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
				},
			},
			ExpectedConfigMapNames: []string{
				"config-map-1",
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
			CurrentState: []*apiv1.ConfigMap{},
			DesiredState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-2",
					},
				},
			},
			ExpectedConfigMapNames: []string{
				"config-map-1",
				"config-map-2",
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
			CurrentState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
				},
			},
			DesiredState:           []*apiv1.ConfigMap{},
			ExpectedConfigMapNames: []string{},
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
			CurrentState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-2",
					},
				},
			},
			DesiredState:           []*apiv1.ConfigMap{},
			ExpectedConfigMapNames: []string{},
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
			CurrentState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-2",
					},
				},
			},
			DesiredState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-2",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-3",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-4",
					},
				},
			},
			ExpectedConfigMapNames: []string{
				"config-map-3",
				"config-map-4",
			},
		},
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := DefaultConfig()
		resourceConfig.CertWatcher = certificatetprtest.NewService()
		resourceConfig.K8sClient = fake.NewSimpleClientset()
		resourceConfig.Logger = microloggertest.New()
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for i, tc := range testCases {
		result, err := newResource.GetCreateState(context.TODO(), tc.Obj, tc.CurrentState, tc.DesiredState)
		if err != nil {
			t.Fatalf("case %d expected %#v got %#v", i+1, nil, err)
		}

		configMaps, ok := result.([]*apiv1.ConfigMap)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*apiv1.ConfigMap{}, result)
		}

		if len(configMaps) != len(tc.ExpectedConfigMapNames) {
			t.Fatalf("case %d expected %d config maps got %d", i+1, len(tc.ExpectedConfigMapNames), len(configMaps))
		}
	}
}

func Test_Resource_CloudConfig_GetDeleteState(t *testing.T) {
	testCases := []struct {
		Obj                    interface{}
		CurrentState           interface{}
		DesiredState           interface{}
		ExpectedConfigMapNames []string
	}{
		// Test 1, in case current state and desired state are empty the delete
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
			CurrentState:           []*apiv1.ConfigMap{},
			DesiredState:           []*apiv1.ConfigMap{},
			ExpectedConfigMapNames: []string{},
		},

		// Test 2, in case current state has one item and equals desired state the
		// delete state should equal the current state.
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
			CurrentState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
				},
			},
			DesiredState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
				},
			},
			ExpectedConfigMapNames: []string{
				"config-map-1",
			},
		},

		// Test 3, in case current state misses one item of desired state the delete
		// state should not contain the missing item of the desired state.
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
			CurrentState: []*apiv1.ConfigMap{},
			DesiredState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
				},
			},
			ExpectedConfigMapNames: []string{},
		},

		// Test 4, in case current state misses items of desired state the delete
		// state should not contain the missing items of the desired state.
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
			CurrentState: []*apiv1.ConfigMap{},
			DesiredState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-2",
					},
				},
			},
			ExpectedConfigMapNames: []string{},
		},

		// Test 5, in case current state contains one item and desired state is
		// empty the delete state should be empty.
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
			CurrentState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
				},
			},
			DesiredState:           []*apiv1.ConfigMap{},
			ExpectedConfigMapNames: []string{},
		},

		// Test 6, in case current state contains items and desired state is empty
		// the delete state should be empty.
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
			CurrentState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-2",
					},
				},
			},
			DesiredState:           []*apiv1.ConfigMap{},
			ExpectedConfigMapNames: []string{},
		},

		// Test 7, in case all items of current state are in desired state and
		// desired state contains more items not being in current state the create
		// state should contain all items being in current state.
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
			CurrentState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-2",
					},
				},
			},
			DesiredState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-2",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-3",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-4",
					},
				},
			},
			ExpectedConfigMapNames: []string{
				"config-map-1",
				"config-map-2",
			},
		},

		// Test 8, in case all items of desired state are in current state and
		// current state contains more items not being in desired state the create
		// state should contain all items being in desired state.
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
			CurrentState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-2",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-3",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-4",
					},
				},
			},
			DesiredState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-2",
					},
				},
			},
			ExpectedConfigMapNames: []string{
				"config-map-1",
				"config-map-2",
			},
		},
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := DefaultConfig()
		resourceConfig.CertWatcher = certificatetprtest.NewService()
		resourceConfig.K8sClient = fake.NewSimpleClientset()
		resourceConfig.Logger = microloggertest.New()
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for i, tc := range testCases {
		result, err := newResource.GetDeleteState(context.TODO(), tc.Obj, tc.CurrentState, tc.DesiredState)
		if err != nil {
			t.Fatalf("case %d expected %#v got %#v", i+1, nil, err)
		}

		configMaps, ok := result.([]*apiv1.ConfigMap)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*apiv1.ConfigMap{}, result)
		}

		if len(configMaps) != len(tc.ExpectedConfigMapNames) {
			t.Fatalf("case %d expected %d config maps got %d", i+1, len(tc.ExpectedConfigMapNames), len(configMaps))
		}
	}
}

func Test_Resource_CloudConfig_GetUpdateState(t *testing.T) {
	testCases := []struct {
		Obj                        interface{}
		CurrentState               interface{}
		DesiredState               interface{}
		SetCtxUpdateAllowed        bool
		SetCtxUpdateNecessary      bool
		ExpectedConfigMapsToCreate []*apiv1.ConfigMap
		ExpectedConfigMapsToDelete []*apiv1.ConfigMap
		ExpectedConfigMapsToUpdate []*apiv1.ConfigMap
		ExpectedCtxUpdateAllowed   bool
		ExpectedCtxUpdateNecessary bool
	}{
		// Test 1, in case current state and desired state are empty the create,
		// delete and update state should be empty.
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
			CurrentState:               []*apiv1.ConfigMap{},
			DesiredState:               []*apiv1.ConfigMap{},
			SetCtxUpdateAllowed:        true,
			SetCtxUpdateNecessary:      false,
			ExpectedConfigMapsToCreate: nil,
			ExpectedConfigMapsToDelete: nil,
			ExpectedConfigMapsToUpdate: nil,
			ExpectedCtxUpdateAllowed:   true,
			ExpectedCtxUpdateNecessary: false,
		},

		// Test 2, in case current state and desired state are equal the create,
		// delete and update state should be empty.
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
			CurrentState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
					Data: map[string]string{
						"key1": "val1",
					},
				},
			},
			DesiredState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
					Data: map[string]string{
						"key1": "val1",
					},
				},
			},
			SetCtxUpdateAllowed:        true,
			SetCtxUpdateNecessary:      false,
			ExpectedConfigMapsToCreate: nil,
			ExpectedConfigMapsToDelete: nil,
			ExpectedConfigMapsToUpdate: nil,
			ExpectedCtxUpdateAllowed:   true,
			ExpectedCtxUpdateNecessary: false,
		},

		// Test 3, in case current state misses one item of desired state the delete
		// state should not contain the missing item of the desired state.
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
			CurrentState: []*apiv1.ConfigMap{},
			DesiredState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
					Data: map[string]string{
						"key1": "val1",
					},
				},
			},
			SetCtxUpdateAllowed:   true,
			SetCtxUpdateNecessary: false,
			ExpectedConfigMapsToCreate: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
					Data: map[string]string{
						"key1": "val1",
					},
				},
			},
			ExpectedConfigMapsToDelete: nil,
			ExpectedConfigMapsToUpdate: nil,
			ExpectedCtxUpdateAllowed:   true,
			ExpectedCtxUpdateNecessary: false,
		},

		// Test 4, in case current state contains two items and desired state is
		// missing one of them the delete state should contain the the missing item
		// from the current state.
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
			CurrentState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
					Data: map[string]string{
						"key1": "val1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-2",
					},
					Data: map[string]string{
						"key2": "val2",
					},
				},
			},
			DesiredState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
					Data: map[string]string{
						"key1": "val1",
					},
				},
			},
			SetCtxUpdateAllowed:        true,
			SetCtxUpdateNecessary:      false,
			ExpectedConfigMapsToCreate: nil,
			ExpectedConfigMapsToDelete: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-2",
					},
					Data: map[string]string{
						"key2": "val2",
					},
				},
			},
			ExpectedConfigMapsToUpdate: nil,
			ExpectedCtxUpdateAllowed:   true,
			ExpectedCtxUpdateNecessary: false,
		},

		// Test 5, in case current state contains two items and desired state is
		// contains the same state but one object is modified internally the update
		// state should contain the the modified item from the current state.
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
			CurrentState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
					Data: map[string]string{
						"key1": "val1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-2",
					},
					Data: map[string]string{
						"key2": "val2-modified",
					},
				},
			},
			DesiredState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
					Data: map[string]string{
						"key1": "val1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-2",
					},
					Data: map[string]string{
						"key2": "val2",
					},
				},
			},
			SetCtxUpdateAllowed:        true,
			SetCtxUpdateNecessary:      false,
			ExpectedConfigMapsToCreate: nil,
			ExpectedConfigMapsToDelete: nil,
			ExpectedConfigMapsToUpdate: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-2",
					},
					Data: map[string]string{
						"key2": "val2",
					},
				},
			},
			ExpectedCtxUpdateAllowed:   true,
			ExpectedCtxUpdateNecessary: true,
		},

		// Test 6, in case current state contains two items and desired state is
		// contains the same state but one object is modified internally the update
		// state should NOT contain the the modified item from the current state
		// when updates are not allowed.
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
			CurrentState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
					Data: map[string]string{
						"key1": "val1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-2",
					},
					Data: map[string]string{
						"key2": "val2-modified",
					},
				},
			},
			DesiredState: []*apiv1.ConfigMap{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-1",
					},
					Data: map[string]string{
						"key1": "val1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "config-map-2",
					},
					Data: map[string]string{
						"key2": "val2",
					},
				},
			},
			SetCtxUpdateAllowed:        false,
			SetCtxUpdateNecessary:      false,
			ExpectedConfigMapsToCreate: nil,
			ExpectedConfigMapsToDelete: nil,
			ExpectedConfigMapsToUpdate: nil,
			ExpectedCtxUpdateAllowed:   false,
			ExpectedCtxUpdateNecessary: false,
		},
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := DefaultConfig()
		resourceConfig.CertWatcher = certificatetprtest.NewService()
		resourceConfig.K8sClient = fake.NewSimpleClientset()
		resourceConfig.Logger = microloggertest.New()
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for i, tc := range testCases {
		ctx := context.Background()

		ctx = updateallowedcontext.NewContext(ctx, make(chan struct{}))
		ctx = updatenecessarycontext.NewContext(ctx, make(chan struct{}))

		if tc.SetCtxUpdateAllowed {
			updateallowedcontext.SetUpdateAllowed(ctx)
		}
		if tc.SetCtxUpdateNecessary {
			updatenecessarycontext.SetUpdateNecessary(ctx)
		}

		createState, deleteState, updateState, err := newResource.GetUpdateState(ctx, tc.Obj, tc.CurrentState, tc.DesiredState)
		if err != nil {
			t.Fatalf("case %d expected %#v got %#v", i+1, nil, err)
		}

		configMapsToCreate, ok := createState.([]*apiv1.ConfigMap)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*apiv1.ConfigMap{}, createState)
		}
		if !reflect.DeepEqual(configMapsToCreate, tc.ExpectedConfigMapsToCreate) {
			t.Fatalf("case %d expected %#v got %#v", i+1, tc.ExpectedConfigMapsToCreate, configMapsToCreate)
		}

		configMapsToDelete, ok := deleteState.([]*apiv1.ConfigMap)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*apiv1.ConfigMap{}, deleteState)
		}
		if !reflect.DeepEqual(configMapsToDelete, tc.ExpectedConfigMapsToDelete) {
			t.Fatalf("case %d expected %#v got %#v", i+1, tc.ExpectedConfigMapsToDelete, configMapsToDelete)
		}

		configMapsToUpdate, ok := updateState.([]*apiv1.ConfigMap)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*apiv1.ConfigMap{}, updateState)
		}
		if !reflect.DeepEqual(configMapsToUpdate, tc.ExpectedConfigMapsToUpdate) {
			t.Fatalf("case %d expected %#v got %#v", i+1, tc.ExpectedConfigMapsToUpdate, configMapsToUpdate)
		}

		isUpdateAllowed := updateallowedcontext.IsUpdateAllowed(ctx)
		if isUpdateAllowed != tc.ExpectedCtxUpdateAllowed {
			t.Fatalf("case %d expected %#v got %#v", i+1, tc.ExpectedCtxUpdateAllowed, isUpdateAllowed)
		}
		isUpdateNecessary := updatenecessarycontext.IsUpdateNecessary(ctx)
		if isUpdateNecessary != tc.ExpectedCtxUpdateNecessary {
			t.Fatalf("case %d expected %#v got %#v", i+1, tc.ExpectedCtxUpdateNecessary, isUpdateNecessary)
		}
	}
}

func testGetMasterCount(configMaps []*apiv1.ConfigMap) int {
	var count int

	for _, c := range configMaps {
		if strings.HasPrefix(c.Name, "master-") {
			count++
		}
	}

	return count
}

func testGetWorkerCount(configMaps []*apiv1.ConfigMap) int {
	var count int

	for _, c := range configMaps {
		if strings.HasPrefix(c.Name, "worker-") {
			count++
		}
	}

	return count
}
