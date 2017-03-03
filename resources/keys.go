package resources

import (
	"fmt"

	"github.com/giantswarm/kvmtpr"
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

func NetworkBridgeName(ID string) string {
	return fmt.Sprintf("br-%s", ID)
}

func NetworkEnvFilePath(ID string) string {
	return fmt.Sprintf("/run/flannel/networks/%s.env", NetworkBridgeName(ID))
}

func CustomerID(customObject kvmtpr.CustomObject) string {
	return customObject.Spec.Cluster.Customer.ID
}
