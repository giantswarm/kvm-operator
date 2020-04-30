package deployment

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	apiextfake "github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_Deployment_newDeleteChange(t *testing.T) {
	// Create a fake release
	release := releasev1alpha1.NewReleaseCR()
	release.ObjectMeta.Name = "v1.0.2"
	release.Spec.Components = []releasev1alpha1.ReleaseSpecComponent{
		{
			Name:    "kubernetes",
			Version: "1.15.11",
		},
		{
			Name:    "calico",
			Version: "3.9.1",
		},
		{
			Name:    "etcd",
			Version: "3.3.15",
		},
	}
	clientset := apiextfake.NewSimpleClientset(release)

	testCases := []struct {
		Obj                     interface{}
		CurrentState            interface{}
		DesiredState            interface{}
		ExpectedDeploymentNames []string
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
			CurrentState:            []*v1.Deployment{},
			DesiredState:            []*v1.Deployment{},
			ExpectedDeploymentNames: []string{},
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
			CurrentState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
			},
			DesiredState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
			},
			ExpectedDeploymentNames: []string{
				"deployment-1",
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
			CurrentState: []*v1.Deployment{},
			DesiredState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
			},
			ExpectedDeploymentNames: []string{},
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
			CurrentState: []*v1.Deployment{},
			DesiredState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
					},
				},
			},
			ExpectedDeploymentNames: []string{},
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
			CurrentState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
			},
			DesiredState:            []*v1.Deployment{},
			ExpectedDeploymentNames: []string{},
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
			CurrentState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
					},
				},
			},
			DesiredState:            []*v1.Deployment{},
			ExpectedDeploymentNames: []string{},
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
			CurrentState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
					},
				},
			},
			DesiredState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-3",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-4",
					},
				},
			},
			ExpectedDeploymentNames: []string{
				"deployment-1",
				"deployment-2",
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
			CurrentState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-3",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-4",
					},
				},
			},
			DesiredState: []*v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-2",
					},
				},
			},
			ExpectedDeploymentNames: []string{
				"deployment-1",
				"deployment-2",
			},
		},
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := Config{
			DNSServers: "dnsserver1,dnsserver2",
			G8sClient:  clientset,
			K8sClient:  fake.NewSimpleClientset(),
			Logger:     microloggertest.New(),
		}
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for i, tc := range testCases {
		result, err := newResource.newDeleteChangeForDeletePatch(context.TODO(), tc.Obj, tc.CurrentState, tc.DesiredState)
		if err != nil {
			t.Fatalf("case %d expected %#v got %#v", i+1, nil, err)
		}

		deployments, ok := result.([]*v1.Deployment)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*v1.Deployment{}, result)
		}

		if len(deployments) != len(tc.ExpectedDeploymentNames) {
			t.Fatalf("case %d expected %d config maps got %d", i+1, len(tc.ExpectedDeploymentNames), len(deployments))
		}
	}
}
