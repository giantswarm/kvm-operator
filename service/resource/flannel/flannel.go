package flannel

import (
	"github.com/giantswarm/kvmtpr"
	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	apiv1 "k8s.io/client-go/pkg/api/v1"
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

// Service implements the flannel service.
type Service struct {
	// Dependencies.
	logger micrologger.Logger
}

func (s *Service) Containers(obj interface{}) ([]apiv1.Container, error) {
	customObject, ok := obj.(*kvmtpr.CustomObject)
	if !ok {
		return nil, microerror.MaskAnyf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, obj)
	}

	return s.newContainers(customObject), nil
}

func (s *Service) InitContainers(obj interface{}) ([]apiv1.Container, error) {
	customObject, ok := obj.(*kvmtpr.CustomObject)
	if !ok {
		return nil, microerror.MaskAnyf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, obj)
	}

	return s.newInitContainers(customObject), nil
}

func (s *Service) Volumes(obj interface{}) ([]apiv1.Volume, error) {
	customObject, ok := obj.(*kvmtpr.CustomObject)
	if !ok {
		return nil, microerror.MaskAnyf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, obj)
	}

	return s.newVolumes(customObject), nil
}
