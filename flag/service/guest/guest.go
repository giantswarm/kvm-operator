package guest

import (
	"github.com/giantswarm/kvm-operator/flag/service/guest/ssh"
	"github.com/giantswarm/kvm-operator/flag/service/guest/update"
)

type Guest struct {
	SSH    ssh.SSH
	Update update.Update
}
