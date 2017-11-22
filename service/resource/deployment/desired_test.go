package deployment

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/giantswarm/clustertpr"
	clustertprspec "github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/kvmtpr"
	kvmtprspec "github.com/giantswarm/kvmtpr/spec"
	kvmtprspeckvm "github.com/giantswarm/kvmtpr/spec/kvm"
	"github.com/giantswarm/kvmtpr/spec/kvm/nodecontroller"
	"github.com/giantswarm/micrologger/microloggertest"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes/fake"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func Test_Resource_Deployment_GetDesiredState(t *testing.T) {
	testCases := []struct {
		Obj                      interface{}
		ExpectedMasterCount      int
		ExpectedNodeCtrlCount    int
		ExpectedWorkerCount      int
		ExpectedMastersResources []apiv1.ResourceRequirements
		ExpectedWorkersResources []apiv1.ResourceRequirements
	}{
		// Test 1 ensures there is one deployment for master and worker each when there
		// is one master and one worker node in the custom object.
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
					KVM: kvmtprspec.KVM{
						K8sKVM: kvmtprspeckvm.K8sKVM{
							StorageType: "hostPath",
						},
						Masters: []kvmtprspeckvm.Node{
							{CPUs: 1, Memory: "1Gi"},
						},
						Workers: []kvmtprspeckvm.Node{
							{CPUs: 4, Memory: "8.2G"},
						},
					},
				},
			},
			ExpectedMasterCount:   1,
			ExpectedNodeCtrlCount: 0,
			ExpectedWorkerCount:   1,
			ExpectedMastersResources: []apiv1.ResourceRequirements{
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("1"),
						apiv1.ResourceMemory: resource.MustParse("1Gi"),
					},
				},
			},
			ExpectedWorkersResources: []apiv1.ResourceRequirements{
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("8.2G"),
					},
				},
			},
		},

		// Test 2 ensures there is one deployment for master and worker each when there
		// is one master and three worker nodes in the custom object.
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
					KVM: kvmtprspec.KVM{
						K8sKVM: kvmtprspeckvm.K8sKVM{
							StorageType: "hostPath",
						},
						Masters: []kvmtprspeckvm.Node{
							{CPUs: 1, Memory: "1Gi"},
						},
						NodeController: kvmtprspeckvm.NodeController{
							Docker: nodecontroller.Docker{
								Image: "123",
							},
						},
						Workers: []kvmtprspeckvm.Node{
							{CPUs: 4, Memory: "8.2G"},
							{CPUs: 4, Memory: "8.2G"},
							{CPUs: 4, Memory: "8.2G"},
						},
					},
				},
			},
			ExpectedMasterCount:   1,
			ExpectedNodeCtrlCount: 1,
			ExpectedWorkerCount:   3,
			ExpectedMastersResources: []apiv1.ResourceRequirements{
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("1"),
						apiv1.ResourceMemory: resource.MustParse("1Gi"),
					},
				},
			},
			ExpectedWorkersResources: []apiv1.ResourceRequirements{
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("8.2G"),
					},
				},
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("8.2G"),
					},
				},
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("8.2G"),
					},
				},
			},
		},

		// Test 3 ensures there is one deployment for master and worker each when there
		// are three master and three worker nodes in the custom object.
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
					KVM: kvmtprspec.KVM{
						K8sKVM: kvmtprspeckvm.K8sKVM{
							StorageType: "hostPath",
						},
						Masters: []kvmtprspeckvm.Node{
							{CPUs: 1, Memory: "1Gi"},
							{CPUs: 1, Memory: "1Gi"},
							{CPUs: 1, Memory: "1Gi"},
						},
						NodeController: kvmtprspeckvm.NodeController{
							Docker: nodecontroller.Docker{
								Image: "123",
							},
						},
						Workers: []kvmtprspeckvm.Node{
							{CPUs: 4, Memory: "8.2G"},
							{CPUs: 4, Memory: "8.2G"},
							{CPUs: 4, Memory: "8.2G"},
						},
					},
				},
			},
			ExpectedMasterCount:   3,
			ExpectedNodeCtrlCount: 1,
			ExpectedWorkerCount:   3,
			ExpectedMastersResources: []apiv1.ResourceRequirements{
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("1"),
						apiv1.ResourceMemory: resource.MustParse("1Gi"),
					},
				},
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("1"),
						apiv1.ResourceMemory: resource.MustParse("1Gi"),
					},
				},
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("1"),
						apiv1.ResourceMemory: resource.MustParse("1Gi"),
					},
				},
			},
			ExpectedWorkersResources: []apiv1.ResourceRequirements{
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("8.2G"),
					},
				},
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("8.2G"),
					},
				},
				{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("8.2G"),
					},
				},
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
		result, err := newResource.GetDesiredState(context.TODO(), tc.Obj)
		if err != nil {
			t.Fatalf("case %d expected %#v got %#v", i+1, nil, err)
		}

		deployments, ok := result.([]*v1beta1.Deployment)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*v1beta1.Deployment{}, result)
		}

		if testGetMasterCount(deployments) != tc.ExpectedMasterCount {
			t.Fatalf("case %d expected %d master nodes got %d", i+1, tc.ExpectedMasterCount, testGetMasterCount(deployments))
		}

		if testGetWorkerCount(deployments) != tc.ExpectedWorkerCount {
			t.Fatalf("case %d expected %d worker nodes got %d", i+1, tc.ExpectedWorkerCount, testGetWorkerCount(deployments))
		}

		if len(deployments) != tc.ExpectedMasterCount+tc.ExpectedWorkerCount+tc.ExpectedNodeCtrlCount {
			t.Fatalf("case %d expected %d nodes got %d", i+1, tc.ExpectedMasterCount+tc.ExpectedWorkerCount+tc.ExpectedNodeCtrlCount, len(deployments))
		}

		if !reflect.DeepEqual(testGetK8sMasterKVMResources(deployments), tc.ExpectedMastersResources) {
			t.Fatalf("case %d expected %#v got %#v", i+1, testGetK8sMasterKVMResources(deployments), tc.ExpectedMastersResources)
		}

		if !reflect.DeepEqual(testGetK8sWorkerKVMResources(deployments), tc.ExpectedWorkersResources) {
			t.Fatalf("case %d expected %#v got %#v", i+1, testGetK8sWorkerKVMResources(deployments), tc.ExpectedWorkersResources)
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
