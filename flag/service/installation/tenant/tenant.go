package tenant

import (
	"github.com/giantswarm/kvm-operator/v4/flag/service/installation/tenant/kubernetes"
)

type Tenant struct {
	Kubernetes kubernetes.Kubernetes
}
