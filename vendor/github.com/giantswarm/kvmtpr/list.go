package kvmtpr

import (
	"k8s.io/client-go/pkg/api/unversioned"
)

// List represents a list of custom objects.
type List struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`

	Items []*CustomObject `json:"items" yaml:"items"`
}
