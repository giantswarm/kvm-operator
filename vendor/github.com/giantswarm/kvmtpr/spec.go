package kvmtpr

import (
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/versionbundle"

	"github.com/giantswarm/kvmtpr/spec"
)

type Spec struct {
	Cluster       clustertpr.Spec      `json:"cluster" yaml:"cluster"`
	KVM           spec.KVM             `json:"kvm" yaml:"kvm"`
	VersionBundle versionbundle.Bundle `json:"version_bundle" yaml:"version_bundle"`
}
