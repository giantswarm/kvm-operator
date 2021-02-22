package workload

import (
	"github.com/giantswarm/kvm-operator/flag/service/installation/workload/kubernetes"
)

type Workload struct {
	Kubernetes kubernetes.Kubernetes
}
