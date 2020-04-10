package configmap

import (
	"context"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	apiextfake "github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/certs/certstest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeys/randomkeystest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/kvm-operator/pkg/label"
	"github.com/giantswarm/kvm-operator/service/controller/cloudconfig/cloudconfigtest"
)

func Test_Resource_CloudConfig_GetDesiredState(t *testing.T) {
	release := releasev1alpha1.NewReleaseCR()
	release.ObjectMeta.Name = "v1.0.0"
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
		Name                string
		Obj                 interface{}
		ExpectedMasterCount int
		ExpectedWorkerCount int
		ErrorMatcher        func(error) bool
	}{
		{
			Name: "single master, single worker",
			Obj: &v1alpha1.KVMConfig{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						label.ReleaseVersion: "1.0.0",
					},
				},
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
						Masters: []v1alpha1.ClusterNode{
							{ID: "a"},
						},
						Workers: []v1alpha1.ClusterNode{
							{ID: "b"},
						},
					},
				},
				Status: v1alpha1.KVMConfigStatus{
					KVM: v1alpha1.KVMConfigStatusKVM{
						NodeIndexes: map[string]int{
							"a": 1,
							"b": 2,
						},
					},
				},
			},
			ExpectedMasterCount: 1,
			ExpectedWorkerCount: 1,
			ErrorMatcher:        nil,
		},
		{
			Name: "single master, three workers",
			Obj: &v1alpha1.KVMConfig{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						label.ReleaseVersion: "1.0.0",
					},
				},
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
						Masters: []v1alpha1.ClusterNode{
							{ID: "a"},
						},
						Workers: []v1alpha1.ClusterNode{
							{ID: "b"},
							{ID: "c"},
							{ID: "d"},
						},
					},
				},
				Status: v1alpha1.KVMConfigStatus{
					KVM: v1alpha1.KVMConfigStatusKVM{
						NodeIndexes: map[string]int{
							"a": 1,
							"b": 2,
							"c": 3,
							"d": 4,
						},
					},
				},
			},
			ExpectedMasterCount: 1,
			ExpectedWorkerCount: 3,
			ErrorMatcher:        nil,
		},
		{
			Name: "three masters, three workers",
			Obj: &v1alpha1.KVMConfig{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						label.ReleaseVersion: "1.0.0",
					},
				},
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
						Masters: []v1alpha1.ClusterNode{
							{ID: "a"},
							{ID: "b"},
							{ID: "c"},
						},
						Workers: []v1alpha1.ClusterNode{
							{ID: "d"},
							{ID: "e"},
							{ID: "f"},
						},
					},
				},
				Status: v1alpha1.KVMConfigStatus{
					KVM: v1alpha1.KVMConfigStatusKVM{
						NodeIndexes: map[string]int{
							"a": 1,
							"b": 2,
							"c": 3,
							"d": 4,
							"e": 5,
							"f": 6,
						},
					},
				},
			},
			ExpectedMasterCount: 3,
			ExpectedWorkerCount: 3,
			ErrorMatcher:        nil,
		},
		{
			Name: "missing node index for worker",
			Obj: &v1alpha1.KVMConfig{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						label.ReleaseVersion: "1.0.0",
					},
				},
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
						Masters: []v1alpha1.ClusterNode{
							{ID: "a"},
						},
						Workers: []v1alpha1.ClusterNode{
							{ID: "b"},
						},
					},
				},
				Status: v1alpha1.KVMConfigStatus{
					KVM: v1alpha1.KVMConfigStatusKVM{
						NodeIndexes: map[string]int{
							"a": 1,
						},
					},
				},
			},
			ExpectedMasterCount: 0,
			ExpectedWorkerCount: 0,
			ErrorMatcher:        IsNotFound,
		},
		{
			Name: "missing node index for master",
			Obj: &v1alpha1.KVMConfig{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						label.ReleaseVersion: "1.0.0",
					},
				},
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
						Masters: []v1alpha1.ClusterNode{
							{ID: "a"},
						},
						Workers: []v1alpha1.ClusterNode{
							{ID: "b"},
						},
					},
				},
				Status: v1alpha1.KVMConfigStatus{
					KVM: v1alpha1.KVMConfigStatusKVM{
						NodeIndexes: map[string]int{
							"b": 1,
						},
					},
				},
			},
			ExpectedMasterCount: 0,
			ExpectedWorkerCount: 0,
			ErrorMatcher:        IsNotFound,
		},
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := Config{}
		resourceConfig.CertsSearcher = certstest.NewSearcher(certstest.Config{})
		resourceConfig.CloudConfig = cloudconfigtest.New()
		resourceConfig.G8sClient = clientset
		resourceConfig.K8sClient = fake.NewSimpleClientset()
		resourceConfig.KeyWatcher = randomkeystest.NewSearcher()
		resourceConfig.Logger = microloggertest.New()
		resourceConfig.RegistryDomain = "example.com"
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			result, err := newResource.GetDesiredState(context.Background(), tc.Obj)

			switch {
			case err == nil && tc.ErrorMatcher == nil:
				// correct; carry on
			case err != nil && tc.ErrorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.ErrorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.ErrorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			case tc.ErrorMatcher(err):
				return
			}

			configMaps, ok := result.([]*corev1.ConfigMap)
			if !ok {
				t.Fatalf("expected %T got %T", []*corev1.ConfigMap{}, result)
			}

			if testGetMasterCount(configMaps) != tc.ExpectedMasterCount {
				t.Fatalf("expected %d master nodes got %d", tc.ExpectedMasterCount, testGetMasterCount(configMaps))
			}

			if testGetWorkerCount(configMaps) != tc.ExpectedWorkerCount {
				t.Fatalf("expected %d worker nodes got %d", tc.ExpectedWorkerCount, testGetWorkerCount(configMaps))
			}

			if len(configMaps) != tc.ExpectedMasterCount+tc.ExpectedWorkerCount {
				t.Fatalf("expected %d nodes got %d", tc.ExpectedMasterCount+tc.ExpectedWorkerCount, len(configMaps))
			}
		})
	}
}

func testGetMasterCount(configMaps []*corev1.ConfigMap) int {
	var count int

	for _, c := range configMaps {
		if strings.HasPrefix(c.Name, "master-") {
			count++
		}
	}

	return count
}

func testGetWorkerCount(configMaps []*corev1.ConfigMap) int {
	var count int

	for _, c := range configMaps {
		if strings.HasPrefix(c.Name, "worker-") {
			count++
		}
	}

	return count
}
