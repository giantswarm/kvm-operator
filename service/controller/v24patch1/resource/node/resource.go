package node

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	"k8s.io/client-go/kubernetes"
)

const (
<<<<<<< HEAD
<<<<<<< HEAD
	Name = "nodev24patch1"
=======
	Name = "nodev24"
>>>>>>> c4c6c79d... copy v24 to v24patch1
=======
	Name = "nodev24patch1"
>>>>>>> d6f149c2... wire v24patch1
)

type Config struct {
	K8sClient     kubernetes.Interface
	Logger        micrologger.Logger
	TenantCluster tenantcluster.Interface
}

type Resource struct {
	k8sClient     kubernetes.Interface
	logger        micrologger.Logger
	tenantCluster tenantcluster.Interface
}

func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.TenantCluster == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.TenantCluster must not be empty", config)
	}

	r := &Resource{
		k8sClient:     config.K8sClient,
		logger:        config.Logger,
		tenantCluster: config.TenantCluster,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
