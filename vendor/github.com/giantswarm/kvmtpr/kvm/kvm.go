package kvm

import (
	"github.com/giantswarm/kvmtpr/kvm/certctl"
	"github.com/giantswarm/kvmtpr/kvm/endpointupdater"
	"github.com/giantswarm/kvmtpr/kvm/k8svm"
	"github.com/giantswarm/kvmtpr/kvm/kubectl"
	"github.com/giantswarm/kvmtpr/kvm/network"
)

type KVM struct {
	Certctl         certctl.Certctl                 `json:"certctl" yaml:"certctl"`
	EndpointUpdater endpointupdater.EndpointUpdater `json:"endpointUpdater" yaml:"endpointUpdater"`
	K8sVM           k8svm.K8sVM                     `json:"k8sVM" yaml:"k8sVM"`
	Kubectl         kubectl.Kubectl                 `json:"kubectl" yaml:"kubectl"`
	Network         network.Network                 `json:"network" yaml:"network"`
}
