package service

import (
	"github.com/giantswarm/kvm-operator/flag/service/guest"
	"github.com/giantswarm/kvm-operator/flag/service/installation"
	"github.com/giantswarm/kvm-operator/flag/service/kubernetes"
)

type Service struct {
	Guest        guest.Guest
	Installation installation.Installation
	Kubernetes   kubernetes.Kubernetes
}
