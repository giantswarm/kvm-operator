package guest

import (
	"github.com/giantswarm/kvm-operator/flag/service/guest/ignition"
	"github.com/giantswarm/kvm-operator/flag/service/guest/ssh"
	"github.com/giantswarm/kvm-operator/flag/service/guest/update"
)

type Guest struct {
	Ignition ignition.Ignition
	SSH      ssh.SSH
	Update   update.Update
}
