package pvc

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_PVC_newCreateChange(t *testing.T) {
	testCases := []struct {
		Obj              interface{}
		CurrentState     interface{}
		DesiredState     interface{}
		ExpectedPVCNames []string
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
			CurrentState:     []corev1.PersistentVolumeClaim{},
			DesiredState:     []corev1.PersistentVolumeClaim{},
			ExpectedPVCNames: []string{},
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
			CurrentState: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
			},
			DesiredState: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
			},
			ExpectedPVCNames: []string{},
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
			CurrentState: []corev1.PersistentVolumeClaim{},
			DesiredState: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
			},
			ExpectedPVCNames: []string{
				"pvc-1",
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
			CurrentState: []corev1.PersistentVolumeClaim{},
			DesiredState: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc-2",
					},
				},
			},
			ExpectedPVCNames: []string{
				"pvc-1",
				"pvc-2",
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
			CurrentState: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
			},
			DesiredState:     []corev1.PersistentVolumeClaim{},
			ExpectedPVCNames: []string{},
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
			CurrentState: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc-2",
					},
				},
			},
			DesiredState:     []corev1.PersistentVolumeClaim{},
			ExpectedPVCNames: []string{},
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
			CurrentState: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc-2",
					},
				},
			},
			DesiredState: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc-2",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc-3",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc-4",
					},
				},
			},
			ExpectedPVCNames: []string{
				"pvc-3",
				"pvc-4",
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

		configMaps, ok := result.([]corev1.PersistentVolumeClaim)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []corev1.PersistentVolumeClaim{}, result)
		}

		if len(configMaps) != len(tc.ExpectedPVCNames) {
			t.Fatalf("case %d expected %d config maps got %d", i+1, len(tc.ExpectedPVCNames), len(configMaps))
		}
	}
}
