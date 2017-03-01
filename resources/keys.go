package resources

import (
	"fmt"

	"github.com/giantswarm/clusterspec"
)

func ClusterCustomer(customObject clusterspec.Cluster) string {
	return customObject.Spec.Customer
}

func ClusterDomain(sub, clusterID, domain string) string {
	return fmt.Sprintf("%s.%s.%s", sub, clusterID, domain)
}

func ClusterID(customObject clusterspec.Cluster) string {
	return customObject.Spec.ClusterId
}

func ClusterName(customObject clusterspec.Cluster) string {
	return ClusterID(customObject)
}

func ClusterNamespace(customObject clusterspec.Cluster) string {
	return ClusterID(customObject)
}

func NetworkBridgeName(ID string) string {
	return fmt.Sprintf("br-%s", ID)
}

func NetworkEnvFilePath(ID string) string {
	return fmt.Sprintf("/run/flannel/networks/%s.env", NetworkBridgeName(ID))
}
