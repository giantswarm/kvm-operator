package terminateunhealthynodes

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "terminateunhealthynodes"
)

// Config represents the configuration used to create a new cloud config resource.
type Config struct {
	// Dependencies.
	K8sClient     kubernetes.Interface
	Logger        micrologger.Logger
	TenantCluster tenantcluster.Interface
}

// Resource implements the config map resource.
type Resource struct {
	// Dependencies.
	k8sClient     kubernetes.Interface
	logger        micrologger.Logger
	tenantCluster tenantcluster.Interface
}

// New creates a new configured config map resource.
func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.TenantCluster == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.TenantCluster must not be empty", config)
	}

	newService := &Resource{
		k8sClient:     config.K8sClient,
		logger:        config.Logger,
		tenantCluster: config.TenantCluster,
	}

	return newService, nil
}

func (r *Resource) Name() string {
	return Name
}
