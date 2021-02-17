package service

import (
	"github.com/giantswarm/operatorkit/v4/pkg/flag/service/kubernetes"

	"github.com/giantswarm/kvm-operator/flag/service/installation"
	"github.com/giantswarm/kvm-operator/flag/service/rbac"
	"github.com/giantswarm/kvm-operator/flag/service/registry"
	"github.com/giantswarm/kvm-operator/flag/service/workload"
)

type Service struct {
	Installation installation.Installation
	Kubernetes   kubernetes.Kubernetes
	RBAC         rbac.RBAC
	Registry     registry.Registry
	Workload     workload.Workload
}
