package kvmtpr

import (
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
)

// CustomObject represents the KVM TPR's custom object. It holds the
// specifications of the resource the KVM operator is interested in.
type CustomObject struct {
	unversioned.TypeMeta `json:",inline"`
	v1.ObjectMeta        `json:"metadata,omitempty"`

	Spec Spec `json:"spec" yaml:"spec"`
}
