package key

import (
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	MasterID         = "master"
	NodeControllerID = "node-controller"
	WorkerID         = "worker"
	// portBase is a baseline for computing the port for liveness probes.
	portBase = 23000
	// HealthEndpoint is http path for liveness probe.
	HealthEndpoint = "/healthz"
	// ProbeHost host for liveness probe.
	ProbeHost = "127.0.0.1"
	// InitialDelaySeconds is InitialDelaySeconds param in liveness probe config
	InitialDelaySeconds = 60
	// TimeoutSeconds is TimeoutSeconds param in liveness probe config
	TimeoutSeconds = 3
	// PeriodSeconds is PeriodSeconds param in liveness probe config
	PeriodSeconds = 20
	// FailureThreshold is FailureThreshold param in liveness probe config
	FailureThreshold = 4
	// SuccessThreshold is SuccessThreshold param in liveness probe config
	SuccessThreshold = 1

	FlannelEnvPathPrefix = "/run/flannel"
	CoreosImageDir       = "/var/lib/coreos-kvm-images"
	CoreosVersion        = "1632.2.1"

	K8SEndpointUpdaterDocker  = "quay.io/giantswarm/k8s-endpoint-updater:df982fc73b71e60fc70a7444c068b52441ddb30e"
	K8SKVMDockerImage         = "quay.io/giantswarm/k8s-kvm:dff62e2b9050b8c16b0fe0a837af113b62db9c0a"
	K8SKVMHealthDocker        = "quay.io/giantswarm/k8s-kvm-health:ddf211dfed52086ade32ab8c45e44eb0273319ef"
	NodeControllerDockerImage = "quay.io/giantswarm/kvm-operator-node-controller:e63163a2fc0a34a519cdd5c787b3e2e4cbadd5a2"

	// defaultWorkerMemory represents the extra memory to add due to qemu overhead.
	defaultWorkerMemory = "1G"

	// kvm endpoint annotations
	AnnotationIp      = "endpoint.kvm.giantswarm.io/ip"
	AnnotationService = "endpoint.kvm.giantswarm.io/service"
)

func ClusterAPIEndpoint(customObject v1alpha1.KVMConfig) string {
	return customObject.Spec.Cluster.Kubernetes.API.Domain
}

func ClusterCustomer(customObject v1alpha1.KVMConfig) string {
	return customObject.Spec.Cluster.Customer.ID
}

func ClusterID(customObject v1alpha1.KVMConfig) string {
	return customObject.Spec.Cluster.ID
}

func ClusterIDFromPod(pod *apiv1.Pod) string {
	l, ok := pod.Labels["cluster"]
	if ok {
		return l
	}

	return "n/a"
}

func ClusterNamespace(customObject v1alpha1.KVMConfig) string {
	return ClusterID(customObject)
}

func ConfigMapName(customObject v1alpha1.KVMConfig, node v1alpha1.ClusterNode, prefix string) string {
	return fmt.Sprintf("%s-%s-%s", prefix, ClusterID(customObject), node.ID)
}

func CPUQuantity(n v1alpha1.KVMConfigSpecKVMNode) (resource.Quantity, error) {
	cpu := strconv.Itoa(n.CPUs)
	q, err := resource.ParseQuantity(cpu)
	if err != nil {
		return resource.Quantity{}, microerror.Mask(err)
	}
	return q, nil
}

func DeploymentName(prefix string, nodeID string) string {
	return fmt.Sprintf("%s-%s", prefix, nodeID)
}

func EtcdPVCName(clusterID string, vmNumber string) string {
	return fmt.Sprintf("%s-%s-%s", "pvc-master-etcd", clusterID, vmNumber)
}

func NetworkEnvFilePath(customObject v1alpha1.KVMConfig) string {
	return fmt.Sprintf("%s/networks/%s.env", FlannelEnvPathPrefix, NetworkBridgeName(customObject))
}

func HealthListenAddress(customObject v1alpha1.KVMConfig) string {
	return "http://" + ProbeHost + ":" + strconv.Itoa(int(LivenessPort(customObject)))
}

func LivenessPort(customObject v1alpha1.KVMConfig) int32 {
	return int32(portBase + customObject.Spec.KVM.Network.Flannel.VNI)
}

func MasterHostPathVolumeDir(clusterID string, vmNumber string) string {
	return filepath.Join("/home/core/volumes", clusterID, "k8s-master-vm"+vmNumber)
}

// MemoryQuantity returns a resource.Quantity that represents the memory to be used by the nodes.
// It adds the memory from the node definition parameter to the additional memory defined by
// defaultWorkerMemory.
func MemoryQuantity(n v1alpha1.KVMConfigSpecKVMNode) (resource.Quantity, error) {
	q, err := resource.ParseQuantity(n.Memory)
	if err != nil {
		return resource.Quantity{}, microerror.Maskf(err, "creating Memory quantity from node definition")
	}
	additionalMemory, err := resource.ParseQuantity(defaultWorkerMemory)
	if err != nil {
		return resource.Quantity{}, microerror.Maskf(err, "creating Memory quantity from addtional memory")
	}
	q.Add(additionalMemory)

	return q, nil
}

func NetworkBridgeName(customObject v1alpha1.KVMConfig) string {
	return fmt.Sprintf("br-%s", ClusterID(customObject))
}

func NetworkTapName(customObject v1alpha1.KVMConfig) string {
	return fmt.Sprintf("tap-%s", ClusterID(customObject))
}

func NetworkDNSBlock(servers []net.IP) string {
	var dnsBlockParts []string

	for _, s := range servers {
		dnsBlockParts = append(dnsBlockParts, fmt.Sprintf("DNS=%s", s.String()))
	}

	dnsBlock := strings.Join(dnsBlockParts, "\n")

	return dnsBlock
}

func NetworkNTPBlock(servers []net.IP) string {
	var ntpBlockParts []string

	for _, s := range servers {
		ntpBlockParts = append(ntpBlockParts, fmt.Sprintf("NTP=%s", s.String()))
	}

	ntpBlock := strings.Join(ntpBlockParts, "\n")

	return ntpBlock
}

func PVCNames(customObject v1alpha1.KVMConfig) []string {
	var names []string

	for i := range customObject.Spec.Cluster.Masters {
		names = append(names, EtcdPVCName(ClusterID(customObject), VMNumber(i)))
	}

	return names
}

func ServiceAccountName(customObject v1alpha1.KVMConfig) string {
	return ClusterID(customObject)
}

func StorageType(customObject v1alpha1.KVMConfig) string {
	return customObject.Spec.KVM.K8sKVM.StorageType
}

func ToCustomObject(v interface{}) (v1alpha1.KVMConfig, error) {
	customObjectPointer, ok := v.(*v1alpha1.KVMConfig)
	if !ok {
		return v1alpha1.KVMConfig{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.KVMConfig{}, v)
	}
	customObject := *customObjectPointer

	return customObject, nil
}

func VersionBundleVersion(customObject v1alpha1.KVMConfig) string {
	return customObject.Spec.VersionBundle.Version
}

func VMNumber(ID int) string {
	return fmt.Sprintf("%d", ID)
}
