package tenant

import (
	"github.com/giantswarm/kvm-operator/flag/service/tenant/ignition"
	"github.com/giantswarm/kvm-operator/flag/service/tenant/ssh"
	"github.com/giantswarm/kvm-operator/flag/service/tenant/update"
)

type Tenant struct {
	Ignition ignition.Ignition
	SSH      ssh.SSH
	Update   update.Update
}
