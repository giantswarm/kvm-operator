package tenant

import (
	"github.com/giantswarm/kvm-operator/flag/service/tenant/docker"
	"github.com/giantswarm/kvm-operator/flag/service/tenant/ignition"
	"github.com/giantswarm/kvm-operator/flag/service/tenant/kubernetes"
	"github.com/giantswarm/kvm-operator/flag/service/tenant/ssh"
	"github.com/giantswarm/kvm-operator/flag/service/tenant/update"
)

type Tenant struct {
	Ignition   ignition.Ignition
	SSH        ssh.SSH
	Update     update.Update
	Docker     docker.Docker
	Kubernetes kubernetes.Kubernetes
}
