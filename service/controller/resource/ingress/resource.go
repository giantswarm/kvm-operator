package ingress

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	APIID  = "api"
	EtcdID = "etcd"
	// Name is the identifier of the resource.
	Name = "ingress"
)

// Config represents the configuration used to create a new ingress resource.
type Config struct {
	// Dependencies.
	CtrlClient client.Client
	Logger     micrologger.Logger
}

// Resource implements the ingress resource.
type Resource struct {
	// Dependencies.
	ctrlClient client.Client
	logger     micrologger.Logger
}

// New creates a new configured ingress resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CtrlClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newResource := &Resource{
		// Dependencies.
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}

func containsIngress(list []*v1beta1.Ingress, item *v1beta1.Ingress) bool {
	for _, l := range list {
		if l.Name == item.Name {
			return true
		}
	}

	return false
}

func toIngresses(v interface{}) ([]*v1beta1.Ingress, error) {
	if v == nil {
		return nil, nil
	}

	ingresses, ok := v.([]*v1beta1.Ingress)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*v1beta1.Ingress{}, v)
	}

	return ingresses, nil
}
