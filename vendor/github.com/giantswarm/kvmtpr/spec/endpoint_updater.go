package spec

import "github.com/giantswarm/kvmtpr/spec/endpointupdater"

type EndpointUpdater struct {
	Docker endpointupdater.Docker `json:"docker" yaml:"docker"`
}
