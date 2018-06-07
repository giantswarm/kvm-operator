package guest

import (
	"github.com/giantswarm/kvm-operator/flag/service/installation/guest/kubernetes"
)

type Guest struct {
	Kubernetes kubernetes.Kubernetes
}
