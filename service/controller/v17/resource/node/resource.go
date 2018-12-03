package node

import (
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

const (
	Name = "nodev16"
)

type Config struct {
	CertsSearcher certs.Interface
	K8sClient     kubernetes.Interface
	Logger        micrologger.Logger
}

type Resource struct {
	certsSearcher certs.Interface
	k8sClient     kubernetes.Interface
	logger        micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.CertsSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CertsSearcher must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		certsSearcher: config.CertsSearcher,
		k8sClient:     config.K8sClient,
		logger:        config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
