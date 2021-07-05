package terminateunhealthynodes

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Name = "terminateunhealthynodes"
)

type Config struct {
	CtrlClient              client.Client
	Logger                  micrologger.Logger
	TenantCluster           tenantcluster.Interface
	TerminateUnhealthyNodes bool
}

type Resource struct {
	ctrlClient              client.Client
	logger                  micrologger.Logger
	tenantCluster           tenantcluster.Interface
	terminateUnhealthyNodes bool
}

func New(config Config) (*Resource, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CtrlClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.TenantCluster == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.TenantCluster must not be empty", config)
	}

	newService := &Resource{
		ctrlClient:              config.CtrlClient,
		logger:                  config.Logger,
		tenantCluster:           config.TenantCluster,
		terminateUnhealthyNodes: config.TerminateUnhealthyNodes,
	}

	return newService, nil
}

func (r *Resource) Name() string {
	return Name
}
