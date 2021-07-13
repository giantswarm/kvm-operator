package workload

import (
	"github.com/giantswarm/kvm-operator/v4/flag/service/installation/workload/kubernetes"
)

type Workload struct {
	Kubernetes kubernetes.Kubernetes
}
