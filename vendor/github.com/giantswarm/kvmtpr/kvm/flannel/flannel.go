package flannel

import (
	"github.com/giantswarm/kvmtpr/kvm/flannel/client"
	"github.com/giantswarm/kvmtpr/kvm/flannel/docker"
)

type Flannel struct {
	// Backend is the Flannel backend type, e.g. vxlan.
	Backend string `json:"backend" yaml:"backend"`
	// Client is the block for the flannel client configuration.
	Client client.Client `json:"client" yaml:"client"`
	// Docker is the block for the full Docker image tag.
	Docker docker.Docker `json:"docker" yaml:"docker"`
	// Interface is the network interface name, e.g. bond0.3, or ens33.
	Interface string `json:"interface" yaml:"interface"`
	// Network is the subnet specification, e.g. 10.0.9.0/16.
	Network string `json:"network" yaml:"network"`
	// PrivateNetwork is the private network to block against, e.g. 10.0.4.0/24.
	PrivateNetwork string `json:"privateNetwork" yaml:"privateNetwork"`
	// VNI is the vxlan network identifier, e.g. 9.
	VNI int `json:"vni" yaml:"vni"`
}
