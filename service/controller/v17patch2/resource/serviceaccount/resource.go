package serviceaccount

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "serviceaccountv17patch2"
)

// Config represents the configuration used to create a new cloud config resource.
type Config struct {
	// Dependencies.
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new config map
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		K8sClient: nil,
		Logger:    nil,
	}
}

// Resource implements the config map resource.
type Resource struct {
	// Dependencies.
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

// New creates a new configured config map resource.
func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newService := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return newService, nil
}

func (r *Resource) Name() string {
	return Name
}

func toServiceAccount(v interface{}) (*apiv1.ServiceAccount, error) {
	if v == nil {
		return nil, nil
	}

	serviceAccount, ok := v.(*apiv1.ServiceAccount)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", apiv1.ServiceAccount{}, v)
	}

	return serviceAccount, nil
}
