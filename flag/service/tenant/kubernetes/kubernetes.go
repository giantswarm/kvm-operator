package kubernetes

import (
	"github.com/giantswarm/kvm-operator/flag/service/tenant/kubernetes/api"
	"github.com/giantswarm/kvm-operator/flag/service/tenant/kubernetes/kubelet"
	"github.com/giantswarm/kvm-operator/flag/service/tenant/kubernetes/networksetup"
)

type Kubernetes struct {
	API api.API
	Kubelet kubelet.Kubelet
	NetworkSetup networksetup.NetworkSetup
}
