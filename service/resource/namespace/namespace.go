package namespace

import (
	"github.com/giantswarm/kvmtpr"
	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"

	"github.com/giantswarm/kvm-operator/resources"
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new service by
// best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger: nil,
	}
}

// New creates a new configured service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "config.Logger must not be empty")
	}

	newService := &Service{
		// Dependencies.
		logger: config.Logger,
	}

	return newService, nil
}

// Service implements the namespace service.
type Service struct {
	// Dependencies.
	logger micrologger.Logger
}

func (s *Service) GetForCreate(obj interface{}) (runtime.Object, error) {
	customObject, ok := obj.(*kvmtpr.CustomObject)
	if !ok {
		return nil, microerror.MaskAnyf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, obj)
	}

	namespace := &v1.Namespace{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: resources.ClusterNamespace(*customObject),
			Labels: map[string]string{
				"cluster":  resources.ClusterID(*customObject),
				"customer": resources.ClusterCustomer(*customObject),
			},
		},
	}

	return namespace, nil
}

func (s *Service) GetForDelete(obj interface{}) (runtime.Object, error) {
	// Deleting a namespace requires the same resource properties being used
	// during creation. E.g. the object meta name has to be equal. This is why we
	// just redirect here.
	ro, err := s.GetForCreate(obj)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	return ro, nil
}
