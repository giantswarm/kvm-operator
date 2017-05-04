package openssl

import (
	"github.com/giantswarm/kvmtpr/kvm/network/openssl/docker"
)

type OpenSSL struct {
	Docker docker.Docker `json:"docker" yaml:"docker"`
}
