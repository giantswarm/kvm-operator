package deployment

import (
	"context"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_Deployment_GetDesiredState(t *testing.T) {
	testCases := []struct {
		Obj                      interface{}
		ExpectedMasterCount      int
		ExpectedWorkerCount      int
		ExpectedMastersResources []apiv1.ResourceRequirements
		ExpectedWorkersResources []apiv1.ResourceRequirements
	}{
		// Test 0 ensures there is one deployment for master and worker each when
		// there is one master and one worker node in the custom object.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
						Masters: []v1alpha1.ClusterNode{
							{},
						},
						Workers: []v1alpha1.ClusterNode{
							{},
						},
					},
					KVM: v1alpha1.KVMConfigSpecKVM{
						K8sKVM: v1alpha1.KVMConfigSpecKVMK8sKVM{
							StorageType: "hostPath",
						},
						Masters: []v1alpha1.KVMConfigSpecKVMNode{
							{CPUs: 1, Memory: "1G"},
						},
						Workers: []v1alpha1.KVMConfigSpecKVMNode{
							{CPUs: 4, Memory: "8G"},
						},
					},
				},
			},
			ExpectedMasterCount: 1,
			ExpectedWorkerCount: 1,
			ExpectedMastersResources: []apiv1.ResourceRequirements{
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("1"),
						apiv1.ResourceMemory: resource.MustParse("2536M"),
					},
					Limits: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("1"),
						apiv1.ResourceMemory: resource.MustParse("2536M"),
					},
				},
			},
			ExpectedWorkersResources: []apiv1.ResourceRequirements{
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("10240M"),
					},
					Limits: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("10240M"),
					},
				},
			},
		},

		// Test 1 ensures there is one deployment for master and worker each when
		// there is one master and three worker nodes in the custom object.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
						Masters: []v1alpha1.ClusterNode{
							{},
						},
						Workers: []v1alpha1.ClusterNode{
							{},
							{},
							{},
						},
					},
					KVM: v1alpha1.KVMConfigSpecKVM{
						K8sKVM: v1alpha1.KVMConfigSpecKVMK8sKVM{
							StorageType: "hostPath",
						},
						Masters: []v1alpha1.KVMConfigSpecKVMNode{
							{CPUs: 1, Memory: "1G"},
						},
						Workers: []v1alpha1.KVMConfigSpecKVMNode{
							{CPUs: 4, Memory: "8G"},
							{CPUs: 4, Memory: "8G"},
							{CPUs: 4, Memory: "8G"},
						},
					},
				},
			},
			ExpectedMasterCount: 1,
			ExpectedWorkerCount: 3,
			ExpectedMastersResources: []apiv1.ResourceRequirements{
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("1"),
						apiv1.ResourceMemory: resource.MustParse("2536M"),
					},
					Limits: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("1"),
						apiv1.ResourceMemory: resource.MustParse("2536M"),
					},
				},
			},
			ExpectedWorkersResources: []apiv1.ResourceRequirements{
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("10240M"),
					},
					Limits: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("10240M"),
					},
				},
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("10240M"),
					},
					Limits: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("10240M"),
					},
				},
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("10240M"),
					},
					Limits: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("10240M"),
					},
				},
			},
		},

		// Test 2 ensures there is one deployment for master and worker each when
		// there are three master and three worker nodes in the custom object.
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
						Masters: []v1alpha1.ClusterNode{
							{},
							{},
							{},
						},
						Workers: []v1alpha1.ClusterNode{
							{},
							{},
							{},
						},
					},
					KVM: v1alpha1.KVMConfigSpecKVM{
						K8sKVM: v1alpha1.KVMConfigSpecKVMK8sKVM{
							StorageType: "hostPath",
						},
						Masters: []v1alpha1.KVMConfigSpecKVMNode{
							{CPUs: 1, Memory: "1G"},
							{CPUs: 1, Memory: "1G"},
							{CPUs: 1, Memory: "1G"},
						},
						Workers: []v1alpha1.KVMConfigSpecKVMNode{
							{CPUs: 4, Memory: "8G"},
							{CPUs: 4, Memory: "8G"},
							{CPUs: 4, Memory: "8G"},
						},
					},
				},
			},
			ExpectedMasterCount: 3,
			ExpectedWorkerCount: 3,
			ExpectedMastersResources: []apiv1.ResourceRequirements{
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("1"),
						apiv1.ResourceMemory: resource.MustParse("2536M"),
					},
					Limits: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("1"),
						apiv1.ResourceMemory: resource.MustParse("2536M"),
					},
				},
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("1"),
						apiv1.ResourceMemory: resource.MustParse("2536M"),
					},
					Limits: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("1"),
						apiv1.ResourceMemory: resource.MustParse("2536M"),
					},
				},
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("1"),
						apiv1.ResourceMemory: resource.MustParse("2536M"),
					},
					Limits: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("1"),
						apiv1.ResourceMemory: resource.MustParse("2536M"),
					},
				},
			},
			ExpectedWorkersResources: []apiv1.ResourceRequirements{
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("10240M"),
					},
					Limits: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("10240M"),
					},
				},
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("10240M"),
					},
					Limits: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("10240M"),
					},
				},
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("10240M"),
					},
					Limits: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("10240M"),
					},
				},
			},
		},
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := DefaultConfig()
		resourceConfig.DNSServers = "dnsserver1,dnsserver2"
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
			t.Fatalf("case %d expected %#v got %#v", i, nil, err)
		}

		deployments, ok := result.([]*v1beta1.Deployment)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i, []*v1beta1.Deployment{}, result)
		}

		if testGetMasterCount(deployments) != tc.ExpectedMasterCount {
			t.Fatalf("case %d expected %d master nodes got %d", i, tc.ExpectedMasterCount, testGetMasterCount(deployments))
		}

		if testGetWorkerCount(deployments) != tc.ExpectedWorkerCount {
			t.Fatalf("case %d expected %d worker nodes got %d", i, tc.ExpectedWorkerCount, testGetWorkerCount(deployments))
		}

		if len(deployments) != tc.ExpectedMasterCount+tc.ExpectedWorkerCount {
			t.Fatalf("case %d expected %d nodes got %d", i, tc.ExpectedMasterCount+tc.ExpectedWorkerCount, len(deployments))
		}

		for j, r := range testGetK8sMasterKVMResources(deployments) {
			expectedCPU := tc.ExpectedMastersResources[j].Requests.Cpu()
			if r.Requests.Cpu().Cmp(*expectedCPU) != 0 {
				t.Fatalf("case %d expected %#v got %#v", i, expectedCPU, r.Requests.Cpu())
			}
			expectedMemory := tc.ExpectedMastersResources[j].Requests.Memory()
			if r.Requests.Memory().Cmp(*expectedMemory) != 0 {
				t.Fatalf("case %d expected %#v got %#v", i, expectedMemory, r.Requests.Memory())
			}
		}

		for j, r := range testGetK8sWorkerKVMResources(deployments) {
			expectedCPU := tc.ExpectedWorkersResources[j].Requests.Cpu()
			if r.Requests.Cpu().Cmp(*expectedCPU) != 0 {
				t.Fatalf("case %d expected %#v got %#v", i, expectedCPU, r.Requests.Cpu())
			}
			expectedMemory := tc.ExpectedWorkersResources[j].Requests.Memory()
			if r.Requests.Memory().Cmp(*expectedMemory) != 0 {
				t.Fatalf("case %d expected %#v got %#v", i, expectedMemory, r.Requests.Memory())
			}
		}
	}
}

func testGetMasterCount(deployments []*v1beta1.Deployment) int {
	return testGetCountPrefix(deployments, "master-")
}

func testGetWorkerCount(deployments []*v1beta1.Deployment) int {
	return testGetCountPrefix(deployments, "worker-")
}

func testGetCountPrefix(deployments []*v1beta1.Deployment, prefix string) int {
	var count int

	for _, d := range deployments {
		if strings.HasPrefix(d.Name, prefix) {
			count++
		}
	}

	return count
}

func testGetK8sMasterKVMResources(deployments []*v1beta1.Deployment) []apiv1.ResourceRequirements {
	return testGetK8sKVMResourcesPrefix(deployments, "master-")
}

func testGetK8sWorkerKVMResources(deployments []*v1beta1.Deployment) []apiv1.ResourceRequirements {
	return testGetK8sKVMResourcesPrefix(deployments, "worker-")
}

func testGetK8sKVMResourcesPrefix(deployments []*v1beta1.Deployment, prefix string) []apiv1.ResourceRequirements {
	var rs []apiv1.ResourceRequirements

	for _, d := range deployments {
		if !strings.HasPrefix(d.Name, prefix) {
			continue
		}
		for _, c := range d.Spec.Template.Spec.Containers {
			if c.Name == "k8s-kvm" {
				rs = append(rs, c.Resources)
			}
		}
	}

	return rs
}
