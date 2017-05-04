package network

import (
	"github.com/giantswarm/kvmtpr/kvm/network/bridge"
	"github.com/giantswarm/kvmtpr/kvm/network/config"
	"github.com/giantswarm/kvmtpr/kvm/network/environment"
	"github.com/giantswarm/kvmtpr/kvm/network/iptables"
	"github.com/giantswarm/kvmtpr/kvm/network/openssl"
)

type Network struct {
	Bridge      bridge.Bridge           `json:"bridge" yaml:"bridge"`
	Config      config.Config           `json:"config" yaml:"config"`
	Environment environment.Environment `json:"environment" yaml:"environment"`
	IPTables    iptables.IPTables       `json:"ipTables" yaml:"ipTables"`
	OpenSSL     openssl.OpenSSL         `json:"openSSL" yaml:"openSSL"`
}
