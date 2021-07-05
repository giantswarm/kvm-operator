package flag

import (
	"github.com/giantswarm/microkit/flag"

	"github.com/giantswarm/kvm-operator/v4/flag/service"
)

type Flag struct {
	Service service.Service
}

func New() *Flag {
	f := &Flag{}
	flag.Init(f)
	return f
}
