package workload

import (
	"github.com/giantswarm/kvm-operator/v4/flag/service/workload/ignition"
	"github.com/giantswarm/kvm-operator/v4/flag/service/workload/proxy"
	"github.com/giantswarm/kvm-operator/v4/flag/service/workload/ssh"
)

type Workload struct {
	Ignition ignition.Ignition
	Proxy    proxy.Proxy
	SSH      ssh.SSH
}
