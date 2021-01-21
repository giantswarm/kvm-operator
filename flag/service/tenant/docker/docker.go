package docker

import (
	"github.com/giantswarm/kvm-operator/flag/service/tenant/docker/daemon"
)

type Docker struct {
	Daemon daemon.Daemon
}
