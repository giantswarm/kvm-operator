package service

import (
	"github.com/giantswarm/operatorkit/v5/pkg/flag/service/kubernetes"

	"github.com/giantswarm/kvm-operator/v4/flag/service/installation"
	"github.com/giantswarm/kvm-operator/v4/flag/service/rbac"
	"github.com/giantswarm/kvm-operator/v4/flag/service/registry"
	"github.com/giantswarm/kvm-operator/v4/flag/service/workload"
)

type Service struct {
	Installation            installation.Installation
	Kubernetes              kubernetes.Kubernetes
	RBAC                    rbac.RBAC
	Registry                registry.Registry
	TerminateUnhealthyNodes string
	Workload                workload.Workload
}
