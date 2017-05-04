package kubectl

import (
	"github.com/giantswarm/kvmtpr/kvm/kubectl/googleapi"
)

type Kubectl struct {
	GoogleAPI googleapi.GoogleAPI `json:"googleAPI" yaml:"googleAPI"`
}
