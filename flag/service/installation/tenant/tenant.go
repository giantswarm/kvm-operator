package tenant

import (
	"github.com/giantswarm/kvm-operator/flag/service/installation/tenant/kubernetes"
)

type Tenant struct {
	Kubernetes kubernetes.Kubernetes
}
