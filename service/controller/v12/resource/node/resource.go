package node

import (
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

const (
	Name = "nodev12"
)

type Config struct {
	CertsSearcher certs.Interface
	CloudProvider cloudprovider.Interface
	K8sClient     kubernetes.Interface
	Logger        micrologger.Logger
}

type Resource struct {
	certsSearcher certs.Interface
	cloudProvider cloudprovider.Interface
	k8sClient     kubernetes.Interface
	logger        micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.CertsSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CertsSearcher must not be empty", config)
	}
	if config.CloudProvider == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CloudProvider must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		certsSearcher: config.CertsSearcher,
		cloudProvider: config.CloudProvider,
		k8sClient:     config.K8sClient,
		logger:        config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
