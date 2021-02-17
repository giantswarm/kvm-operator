package namespace

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Name is the identifier of the resource.
	Name = "namespace"
)

// Config represents the configuration used to create a new cloud config resource.
type Config struct {
	// Dependencies.
	CtrlClient client.Client
	Logger     micrologger.Logger
}

// Resource implements the cloud config resource.
type Resource struct {
	// Dependencies.
	ctrlClient client.Client
	logger     micrologger.Logger
}

// New creates a new configured cloud config resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CtrlClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newService := &Resource{
		// Dependencies.
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return newService, nil
}

func (r *Resource) Name() string {
	return Name
}

func toNamespace(v interface{}) (*corev1.Namespace, error) {
	if v == nil {
		return nil, nil
	}

	namespace, ok := v.(*corev1.Namespace)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &corev1.Namespace{}, v)
	}

	return namespace, nil
}
