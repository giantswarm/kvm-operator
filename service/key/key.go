package key

import (
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/kvmtpr"
	kvmtprkvm "github.com/giantswarm/kvmtpr/spec/kvm"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	MasterID         = "master"
	NodeControllerID = "node-controller"
	WorkerID         = "worker"
)

func ClusterAPIEndpoint(customObject kvmtpr.CustomObject) string {
	return customObject.Spec.Cluster.Kubernetes.API.Domain
}

func ClusterCustomer(customObject kvmtpr.CustomObject) string {
	return customObject.Spec.Cluster.Customer.ID
}

func ClusterID(customObject kvmtpr.CustomObject) string {
	return customObject.Spec.Cluster.Cluster.ID
}

func ClusterNamespace(customObject kvmtpr.CustomObject) string {
	return ClusterID(customObject)
}

func ConfigMapName(customObject kvmtpr.CustomObject, node spec.Node, prefix string) string {
	return fmt.Sprintf("%s-%s-%s", prefix, ClusterID(customObject), node.ID)
}

func CPUQuantity(n kvmtprkvm.Node) (resource.Quantity, error) {
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

func HasNodeController(customObject kvmtpr.CustomObject) bool {
	if customObject.Spec.KVM.NodeController != (kvmtprkvm.NodeController{}) {
		return true
	}
	return false
}

func MasterHostPathVolumeDir(clusterID string, vmNumber string) string {
	return filepath.Join("/home/core/volumes", clusterID, "k8s-master-vm"+vmNumber)
}

func MemoryQuantity(n kvmtprkvm.Node) (resource.Quantity, error) {
	q, err := resource.ParseQuantity(n.Memory)
	if err != nil {
		return resource.Quantity{}, microerror.Mask(err)
	}
	return q, nil
}

func NetworkBridgeName(customObject kvmtpr.CustomObject) string {
	return fmt.Sprintf("br-%s", ClusterID(customObject))
}

func NetworkTapName(customObject kvmtpr.CustomObject) string {
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

func PVCNames(customObject kvmtpr.CustomObject) []string {
	var names []string

	for i := range customObject.Spec.Cluster.Masters {
		names = append(names, EtcdPVCName(ClusterID(customObject), VMNumber(i)))
	}

	return names
}

func StorageType(customObject kvmtpr.CustomObject) string {
	return customObject.Spec.KVM.K8sKVM.StorageType
}

func ToCustomObject(v interface{}) (kvmtpr.CustomObject, error) {
	customObjectPointer, ok := v.(*kvmtpr.CustomObject)
	if !ok {
		return kvmtpr.CustomObject{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, v)
	}
	customObject := *customObjectPointer

	return customObject, nil
}

func VersionBundleVersion(customObject kvmtpr.CustomObject) string {
	return customObject.Spec.VersionBundle.Version
}

func VMNumber(ID int) string {
	return fmt.Sprintf("%d", ID)
}
