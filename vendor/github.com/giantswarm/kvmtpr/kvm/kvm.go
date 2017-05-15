package kvm

import (
	"github.com/giantswarm/kvmtpr/kvm/certctl"
	"github.com/giantswarm/kvmtpr/kvm/endpointupdater"
	"github.com/giantswarm/kvmtpr/kvm/k8skvm"
	"github.com/giantswarm/kvmtpr/kvm/kubectl"
	"github.com/giantswarm/kvmtpr/kvm/network"
)

type KVM struct {
	Certctl         certctl.Certctl                 `json:"certctl" yaml:"certctl"`
	EndpointUpdater endpointupdater.EndpointUpdater `json:"endpointUpdater" yaml:"endpointUpdater"`
	K8sKVM          k8skvm.K8sKVM                   `json:"k8sKVM" yaml:"k8sKVM"`
	Kubectl         kubectl.Kubectl                 `json:"kubectl" yaml:"kubectl"`
	Network         network.Network                 `json:"network" yaml:"network"`
}
