package namespace

import (
	"github.com/giantswarm/kvmtpr"
	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"

	"github.com/giantswarm/kvm-operator/service/resource"
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

// GetForCreate returns the Kubernetes runtime object for the namespace resource
// being used in reconciliation loops on create events.
func (s *Service) GetForCreate(obj interface{}) ([]runtime.Object, error) {
	ros, err := s.newRuntimeObjects(obj)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	return ros, nil
}

// GetForDelete returns the Kubernetes runtime object for the namespace resource
// being used in reconciliation loops on delete events. Note that deleting a
// namespace requires the same resource properties being used during creation.
// E.g. the object meta name has to be equal. This is why we just use the same
// runtime object for deletion that we already used for creation.
func (s *Service) GetForDelete(obj interface{}) ([]runtime.Object, error) {
	ros, err := s.newRuntimeObjects(obj)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	return ros, nil
}

func (s *Service) newRuntimeObjects(obj interface{}) ([]runtime.Object, error) {
	var runtimeObjects []runtime.Object

	customObject, ok := obj.(*kvmtpr.CustomObject)
	if !ok {
		return nil, microerror.MaskAnyf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, obj)
	}

	newNamespace := &v1.Namespace{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: resource.ClusterNamespace(*customObject),
			Labels: map[string]string{
				"cluster":  resource.ClusterID(*customObject),
				"customer": resource.ClusterCustomer(*customObject),
			},
		},
	}

	runtimeObjects = append(runtimeObjects, newNamespace)

	return runtimeObjects, nil
}
