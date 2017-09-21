package key

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/microerror"
)

const (
	PrefixMaster = "master"
	PrefixWorker = "worker"
)

func ClusterCustomer(customObject kvmtpr.CustomObject) string {
	return customObject.Spec.Cluster.Customer.ID
}

func ClusterDomain(sub, clusterID, domain string) string {
	return fmt.Sprintf("%s.%s.%s", sub, clusterID, domain)
}

func ClusterID(customObject kvmtpr.CustomObject) string {
	return customObject.Spec.Cluster.Cluster.ID
}

func ClusterName(customObject kvmtpr.CustomObject) string {
	return ClusterID(customObject)
}

func ClusterNamespace(customObject kvmtpr.CustomObject) string {
	return ClusterID(customObject)
}

func ConfigMapName(customObject kvmtpr.CustomObject, node spec.Node, prefix string) string {
	return fmt.Sprintf("%s-%s-%s", prefix, ClusterID(customObject), node.ID)
}

func ConfigMapNames(customObject kvmtpr.CustomObject) []string {
	var names []string

	for _, node := range customObject.Spec.Cluster.Masters {
		name := ConfigMapName(customObject, node, PrefixMaster)
		names = append(names, name)
	}

	for _, node := range customObject.Spec.Cluster.Workers {
		name := ConfigMapName(customObject, node, PrefixWorker)
		names = append(names, name)
	}

	return names
}

func DeploymentName(prefix string, nodeID string) string {
	return fmt.Sprintf("%s-%s", prefix, nodeID)
}

func EtcdPVCName(clusterID string, vmNumber string) string {
	return fmt.Sprintf("%s-%s-%s", "pvc-master-etcd", clusterID, vmNumber)
}

func MasterHostPathVolumeDir(clusterID string, vmNumber string) string {
	return filepath.Join("/home/core/volumes", clusterID, "k8s-master-vm"+vmNumber)
}

func NetworkBridgeName(ID string) string {
	return fmt.Sprintf("br-%s", ID)
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

func NetworkEnvFilePath(ID string) string {
	return fmt.Sprintf("/run/flannel/networks/%s.env", NetworkBridgeName(ID))
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

func VMNumber(ID int) string {
	return fmt.Sprintf("%d", ID)
}
