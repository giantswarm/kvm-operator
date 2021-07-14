package deployment

import (
	"context"
	"testing"
	"time"

	"github.com/giantswarm/apiextensions/v3/pkg/clientset/versioned/scheme"
	"github.com/giantswarm/certs/v3/pkg/certs"
	workloadcluster "github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	ctrlfake "sigs.k8s.io/controller-runtime/pkg/client/fake" //nolint

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_Deployment_newCreateChange(t *testing.T) {
	// Create a fake release
	release := releasev1alpha1.NewReleaseCR()
	release.ObjectMeta.Name = "v1.0.1"
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

	testCases := []struct {
		Obj                     interface{}
		CurrentState            interface{}
		DesiredState            interface{}
		ExpectedDeploymentNames []string
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
			CurrentState:            []*v1.Deployment{},
			DesiredState:            []*v1.Deployment{},
			ExpectedDeploymentNames: []string{},
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
			ExpectedDeploymentNames: []string{},
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
			CurrentState: []*v1.Deployment{},
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
			ExpectedDeploymentNames: []string{
				"deployment-1",
				"deployment-2",
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
				"deployment-3",
				"deployment-4",
			},
		},
	}

	logger := microloggertest.New()

	var err error
	var certsSearcher certs.Interface
	{
		c := certs.Config{
			K8sClient:    k8sfake.NewSimpleClientset(),
			Logger:       logger,
			WatchTimeout: 5 * time.Second,
		}

		certsSearcher, err = certs.NewSearcher(c)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	var workloadCluster workloadcluster.Interface
	{
		c := workloadcluster.Config{
			CertsSearcher: certsSearcher,
			Logger:        logger,
			CertID:        certs.APICert,
		}

		workloadCluster, err = workloadcluster.New(c)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	var newResource *Resource
	{
		resourceConfig := Config{
			DNSServers:      "dnsserver1,dnsserver2",
			CtrlClient:      ctrlfake.NewFakeClientWithScheme(scheme.Scheme, release),
			Logger:          logger,
			WorkloadCluster: workloadCluster,
		}
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

		deployments, ok := result.([]*v1.Deployment)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*v1.Deployment{}, result)
		}

		if len(deployments) != len(tc.ExpectedDeploymentNames) {
			t.Fatalf("case %d expected %d config maps got %d", i+1, len(tc.ExpectedDeploymentNames), len(deployments))
		}
	}
}
