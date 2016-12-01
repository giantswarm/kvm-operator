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
	NumNodes int32 `json:"numNodes,omitempty"`
}

func (c *Cluster) GetObjectKind() unversioned.ObjectKind {
	return nil
}
