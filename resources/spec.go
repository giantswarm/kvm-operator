package resources

import (
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
)

type ClusterList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`

	Items []*Cluster `json:"items"`
}

type Cluster struct {
	unversioned.TypeMeta `json:",inline"`
	v1.ObjectMeta        `json:"metadata,omitempty"`
	Spec                 ClusterSpec `json:"spec"`
}

type ClusterSpec struct {
	Customer  string `json:"customer"`
	ClusterID string `json:"clusterId"`
	Replicas  int32  `json:"replicas,omitempty"`
	NumNodes  int32  `json:"numNodes,omitempty"`

	ClusterVNI int32 `json:"clusterVNI,omitempty"`
	ClusterBackend string `json:"clusterBackend"`
	ClusterNetwork string `json:"clusterNetwork"`

	DockerExtraArgs string `json:"dockerExtraArgs,omitempty"`

	Registry string `json:"registry"`

	MachineMem string `json:"machineMem"`
	MachineCPUcores int32 `json:"machineCPUcores"`

	K8sAPIaltNames string `json:"k8sAPIaltNames"`
	K8sVersion string `json:"k8sVersion"`
	K8sDomain string `json:"k8sDomain"`
	K8sETCDdomainName string `json:"k8sETCDdomainName"`
	K8sMasterDomainName string `json:"k8sMasterDomainName"`
	K8sMasterServiceName string `json:"k8sMasterServiceName"`
	K8sNodeLabels string `json:"k8sNodeLabels"`

	VaultToken string `json:"vaulToken"`
}

func (c *Cluster) GetObjectKind() unversioned.ObjectKind {
	return nil
}
