package workload

import (
	"github.com/giantswarm/kvm-operator/flag/service/workload/ignition"
	"github.com/giantswarm/kvm-operator/flag/service/workload/ssh"
)

type Workload struct {
	Ignition ignition.Ignition
	SSH      ssh.SSH
}
