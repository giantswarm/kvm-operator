package service

import (
	"github.com/giantswarm/operatorkit/v5/pkg/flag/service/kubernetes"

	"github.com/giantswarm/kvm-operator/v4/flag/service/crd"
	"github.com/giantswarm/kvm-operator/v4/flag/service/installation"
	"github.com/giantswarm/kvm-operator/v4/flag/service/rbac"
	"github.com/giantswarm/kvm-operator/v4/flag/service/registry"
	"github.com/giantswarm/kvm-operator/v4/flag/service/tenant"
)

type Service struct {
	CRD                     crd.CRD
	Installation            installation.Installation
	Kubernetes              kubernetes.Kubernetes
	RBAC                    rbac.RBAC
	Registry                registry.Registry
	Tenant                  tenant.Tenant
	TerminateUnhealthyNodes string
}
