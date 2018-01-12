package configmapv4

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/certstest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeys/randomkeystest"
	apiv1 "k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/kvm-operator/service/cloudconfigv4/cloudconfigtest"
)

func Test_Resource_CloudConfig_newCreateChange(t *testing.T) {
	testCases := []struct {
		Obj                    interface{}
		CurrentState           interface{}
		DesiredState           interface{}
		ExpectedConfigMapNames []string
	}{
		// Test 1, in case current state and desired state are empty the create
		// state should be empty.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
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
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
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
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
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
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
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
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
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
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
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
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
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
		resourceConfig.CertSearcher = certstest.NewSearcher()
		resourceConfig.CloudConfig = cloudconfigtest.New()
		resourceConfig.K8sClient = fake.NewSimpleClientset()
		resourceConfig.KeyWatcher = randomkeystest.NewSearcher()
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

		configMaps, ok := result.([]*apiv1.ConfigMap)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*apiv1.ConfigMap{}, result)
		}

		if len(configMaps) != len(tc.ExpectedConfigMapNames) {
			t.Fatalf("case %d expected %d config maps got %d", i+1, len(tc.ExpectedConfigMapNames), len(configMaps))
		}
	}
}
