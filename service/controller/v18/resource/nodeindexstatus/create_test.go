package nodeindexstatus

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_EnsureCreated(t *testing.T) {
	testCases := []struct {
		name              string
		inputKVMConfig    *v1alpha1.KVMConfig
		expectedKVMConfig *v1alpha1.KVMConfig
		errorMatcher      func(error) bool
	}{
		{
			name: "case 0: no status, add indexes for nodes",
			inputKVMConfig: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						Masters: []v1alpha1.ClusterNode{
							{
								ID: "1dkfa",
							},
						},
						Workers: []v1alpha1.ClusterNode{
							{
								ID: "1ddqd",
							},
							{
								ID: "1dcli",
							},
							{
								ID: "1dlev",
							},
						},
					},
				},
			},
			expectedKVMConfig: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						Masters: []v1alpha1.ClusterNode{
							{
								ID: "1dkfa",
							},
						},
						Workers: []v1alpha1.ClusterNode{
							{
								ID: "1ddqd",
							},
							{
								ID: "1dcli",
							},
							{
								ID: "1dlev",
							},
						},
					},
				},
				Status: v1alpha1.KVMConfigStatus{
					KVM: v1alpha1.KVMConfigStatusKVM{
						NodeIndexes: map[string]int{
							"1dkfa": 1,
							"1ddqd": 2,
							"1dcli": 3,
							"1dlev": 4,
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name: "case 1: existing nodeindexes, add new node node",
			inputKVMConfig: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						Masters: []v1alpha1.ClusterNode{
							{
								ID: "1dkfa",
							},
						},
						Workers: []v1alpha1.ClusterNode{
							{
								ID: "1ddqd",
							},
							{
								ID: "1dcli",
							},
							{
								ID: "1dlev",
							},
							{
								ID: "1dmus",
							},
						},
					},
				},
				Status: v1alpha1.KVMConfigStatus{
					KVM: v1alpha1.KVMConfigStatusKVM{
						NodeIndexes: map[string]int{
							"1dkfa": 1,
							"1ddqd": 2,
							"1dcli": 3,
							"1dlev": 4,
						},
					},
				},
			},
			expectedKVMConfig: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						Masters: []v1alpha1.ClusterNode{
							{
								ID: "1dkfa",
							},
						},
						Workers: []v1alpha1.ClusterNode{
							{
								ID: "1ddqd",
							},
							{
								ID: "1dcli",
							},
							{
								ID: "1dlev",
							},
							{
								ID: "1dmus",
							},
						},
					},
				},
				Status: v1alpha1.KVMConfigStatus{
					KVM: v1alpha1.KVMConfigStatusKVM{
						NodeIndexes: map[string]int{
							"1dkfa": 1,
							"1ddqd": 2,
							"1dcli": 3,
							"1dlev": 4,
							"1dmus": 5,
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name: "case 2: existing nodeindexes with a whole, add two new node nodes",
			inputKVMConfig: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						Masters: []v1alpha1.ClusterNode{
							{
								ID: "1dkfa",
							},
						},
						Workers: []v1alpha1.ClusterNode{
							{
								ID: "1ddqd",
							},
							{
								ID: "1dcli",
							},
							{
								ID: "1dlev",
							},
							{
								ID: "1dmus",
							},
							{
								ID: "1dmyp",
							},
						},
					},
				},
				Status: v1alpha1.KVMConfigStatus{
					KVM: v1alpha1.KVMConfigStatusKVM{
						NodeIndexes: map[string]int{
							"1dkfa": 1,
							"1ddqd": 2,
							"1dcli": 4,
							"1dlev": 5,
						},
					},
				},
			},
			expectedKVMConfig: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						Masters: []v1alpha1.ClusterNode{
							{
								ID: "1dkfa",
							},
						},
						Workers: []v1alpha1.ClusterNode{
							{
								ID: "1ddqd",
							},
							{
								ID: "1dcli",
							},
							{
								ID: "1dlev",
							},
							{
								ID: "1dmus",
							},
							{
								ID: "1dmyp",
							},
						},
					},
				},
				Status: v1alpha1.KVMConfigStatus{
					KVM: v1alpha1.KVMConfigStatusKVM{
						NodeIndexes: map[string]int{
							"1dkfa": 1,
							"1ddqd": 2,
							"1dmus": 3,
							"1dcli": 4,
							"1dlev": 5,
							"1dmyp": 6,
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name: "case 3: existing nodeindexes, remove two nodes",
			inputKVMConfig: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						Masters: []v1alpha1.ClusterNode{
							{
								ID: "1dkfa",
							},
						},
						Workers: []v1alpha1.ClusterNode{
							{
								ID: "1ddqd",
							},
							{
								ID: "1dlev",
							},
							{
								ID: "1dmus",
							},
						},
					},
				},
				Status: v1alpha1.KVMConfigStatus{
					KVM: v1alpha1.KVMConfigStatusKVM{
						NodeIndexes: map[string]int{
							"1dkfa": 1,
							"1ddqd": 2,
							"1dmus": 3,
							"1dcli": 4,
							"1dlev": 5,
							"1dmyp": 6,
						},
					},
				},
			},
			expectedKVMConfig: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						Masters: []v1alpha1.ClusterNode{
							{
								ID: "1dkfa",
							},
						},
						Workers: []v1alpha1.ClusterNode{
							{
								ID: "1ddqd",
							},
							{
								ID: "1dlev",
							},
							{
								ID: "1dmus",
							},
						},
					},
				},
				Status: v1alpha1.KVMConfigStatus{
					KVM: v1alpha1.KVMConfigStatusKVM{
						NodeIndexes: map[string]int{
							"1dkfa": 1,
							"1ddqd": 2,
							"1dmus": 3,
							"1dlev": 5,
						},
					},
				},
			},
			errorMatcher: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var r *Resource
			{
				c := Config{
					G8sClient: fake.NewSimpleClientset(tc.inputKVMConfig),
					Logger:    microloggertest.New(),
				}

				var err error
				r, err = New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			err := r.EnsureCreated(context.Background(), tc.inputKVMConfig)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			kvmConfig, err := r.g8sClient.ProviderV1alpha1().KVMConfigs(tc.inputKVMConfig.GetNamespace()).Get(tc.inputKVMConfig.GetName(), metav1.GetOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(kvmConfig, tc.expectedKVMConfig); diff != "" {
				t.Fatalf("expectations not met after EnsureCreated(): (-got +expected)\n%s\n", diff)
			}
		})
	}
}

func Test_allocateIndex(t *testing.T) {
	testCases := []struct {
		name               string
		lst                []int
		expectedAllocation int
		expectedLst        []int
	}{
		{
			name:               "case 0: allocate first index",
			lst:                nil,
			expectedAllocation: 1,
			expectedLst:        []int{1},
		},
		{
			name:               "case 1: allocate second (sequential) index",
			lst:                []int{1},
			expectedAllocation: 2,
			expectedLst:        []int{1, 2},
		},
		{
			name:               "case 2: allocate third (sequential) index",
			lst:                []int{1, 2},
			expectedAllocation: 3,
			expectedLst:        []int{1, 2, 3},
		},
		{
			name:               "case 3: allocate second (non-sequential) index",
			lst:                []int{3},
			expectedAllocation: 1,
			expectedLst:        []int{1, 3},
		},
		{
			name:               "case 2: allocate third (non-sequential) index",
			lst:                []int{1, 3},
			expectedAllocation: 2,
			expectedLst:        []int{1, 2, 3},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lst, idx := allocateIndex(tc.lst)

			if idx != tc.expectedAllocation {
				t.Fatalf("allocateIndex(%+v) got %d, expected %d", tc.lst, idx, tc.expectedAllocation)
			}

			if diff := cmp.Diff(lst, tc.expectedLst); diff != "" {
				t.Fatalf("allocateIndex(%+v) allocations differ from expected: (-got +expected)\n%s\n", tc.lst, diff)
			}
		})
	}
}
