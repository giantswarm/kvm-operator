package tenant

import (
	"github.com/giantswarm/kvm-operator/flag/service/tenant/ignition"
	"github.com/giantswarm/kvm-operator/flag/service/tenant/proxy"
	"github.com/giantswarm/kvm-operator/flag/service/tenant/ssh"
	"github.com/giantswarm/kvm-operator/flag/service/tenant/update"
)

type Tenant struct {
	Ignition ignition.Ignition
	Proxy    proxy.Proxy
	SSH      ssh.SSH
	Update   update.Update
}
