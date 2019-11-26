package configmap

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/certstest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeys/randomkeystest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/kvm-operator/service/controller/v20/cloudconfig/cloudconfigtest"
)

func Test_Resource_CloudConfig_newUpdateChange(t *testing.T) {
	testCases := []struct {
		Ctx                                  context.Context
		Obj                                  interface{}
		CurrentState                         interface{}
		DesiredState                         interface{}
		ExpectedConfigMapsToUpdate           []*corev1.ConfigMap
		ExpectedMessageContextConfigMapNames []string
	}{
		// Test 0, in case current state and desired state are empty the update
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
			CurrentState:                         []*corev1.ConfigMap{},
			DesiredState:                         []*corev1.ConfigMap{},
			ExpectedConfigMapsToUpdate:           nil,
			ExpectedMessageContextConfigMapNames: nil,
		},

		// Test 1, in case current state and desired state are equal the update
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
			CurrentState: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "config-map-1",
					},
					Data: map[string]string{
						"key1": "val1",
					},
				},
			},
			DesiredState: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
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

		// Test 2, in case current state contains two items and desired state is
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
			CurrentState: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "config-map-1",
					},
					Data: map[string]string{
						"key1": "val1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "config-map-2",
					},
					Data: map[string]string{
						"key2": "val2-modified",
					},
				},
			},
			DesiredState: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "config-map-1",
					},
					Data: map[string]string{
						"key1": "val1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "config-map-2",
					},
					Data: map[string]string{
						"key2": "val2",
					},
				},
			},
			ExpectedConfigMapsToUpdate: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "config-map-2",
					},
					Data: map[string]string{
						"key2": "val2",
					},
				},
			},
			ExpectedMessageContextConfigMapNames: nil,
		},
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := Config{}
		resourceConfig.CertsSearcher = certstest.NewSearcher(certstest.Config{})
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
		updateState, err := newResource.newUpdateChange(tc.Ctx, tc.Obj, tc.CurrentState, tc.DesiredState)
		if err != nil {
			t.Fatalf("case %d expected %#v got %#v", i, nil, err)
		}

		configMapsToUpdate, ok := updateState.([]*corev1.ConfigMap)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i, []*corev1.ConfigMap{}, updateState)
		}
		if !reflect.DeepEqual(configMapsToUpdate, tc.ExpectedConfigMapsToUpdate) {
			t.Fatalf("case %d expected %#v got %#v", i, tc.ExpectedConfigMapsToUpdate, configMapsToUpdate)
		}
	}
}
