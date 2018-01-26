package configmapv2

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/certstest"
	"github.com/giantswarm/micrologger/microloggertest"
	apiv1 "k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/kvm-operator/service/kvmconfig/cloudconfigv2/cloudconfigtest"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/messagecontext"
)

func Test_Resource_CloudConfig_newUpdateChange(t *testing.T) {
	testCases := []struct {
		Ctx                                  context.Context
		Obj                                  interface{}
		CurrentState                         interface{}
		DesiredState                         interface{}
		ExpectedConfigMapsToUpdate           []*apiv1.ConfigMap
		ExpectedMessageContextConfigMapNames []string
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
			CurrentState:                         []*apiv1.ConfigMap{},
			DesiredState:                         []*apiv1.ConfigMap{},
			ExpectedConfigMapsToUpdate:           nil,
			ExpectedMessageContextConfigMapNames: nil,
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
			ExpectedConfigMapsToUpdate:           nil,
			ExpectedMessageContextConfigMapNames: nil,
		},

		// Test 3, in case current state contains two items and desired state is
		// contains the same state but one object is modified internally the update
		// state should contain the the modified item from the current state.
		{
			Ctx: context.TODO(),
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
			ExpectedMessageContextConfigMapNames: nil,
		},

		// Test 4, same as 3 but ensuring the right config map names are written to
		// the message context.
		{
			Ctx: messagecontext.NewContext(context.Background(), messagecontext.NewMessage()),
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
			ExpectedMessageContextConfigMapNames: []string{
				"config-map-2",
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
		resourceConfig.Logger = microloggertest.New()
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for i, tc := range testCases {
		updateState, err := newResource.newUpdateChange(tc.Ctx, tc.Obj, tc.CurrentState, tc.DesiredState)
		if err != nil {
			t.Fatalf("case %d expected %#v got %#v", i+1, nil, err)
		}

		configMapsToUpdate, ok := updateState.([]*apiv1.ConfigMap)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*apiv1.ConfigMap{}, updateState)
		}
		if !reflect.DeepEqual(configMapsToUpdate, tc.ExpectedConfigMapsToUpdate) {
			t.Fatalf("case %d expected %#v got %#v", i+1, tc.ExpectedConfigMapsToUpdate, configMapsToUpdate)
		}

		m, ok := messagecontext.FromContext(tc.Ctx)
		if ok {
			if !reflect.DeepEqual(m.ConfigMapNames, tc.ExpectedMessageContextConfigMapNames) {
				t.Fatalf("case %d expected %#v got %#v", i+1, tc.ExpectedMessageContextConfigMapNames, m.ConfigMapNames)
			}
		}
	}
}
