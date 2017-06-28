package master

import (
	"github.com/giantswarm/kvm-operator/service/resource/flannel"
	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	"k8s.io/apimachinery/pkg/runtime"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	extensionsv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	Logger  micrologger.Logger
	Flannel *flannel.Service
}

// DefaultConfig provides a default configuration to create a new service by
// best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger:  nil,
		Flannel: nil,
	}
}

// New creates a new configured service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.Flannel == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "config.Flannel must not be empty")
	}

	newService := &Service{
		// Dependencies.
		logger:  config.Logger,
		flannel: config.Flannel,
	}

	return newService, nil
}

// Service implements the master service.
type Service struct {
	// Dependencies.
	logger  micrologger.Logger
	flannel *flannel.Service
}

// GetForCreate returns the Kubernetes runtime object for the master resource
// being used in reconciliation loops on create events.
func (s *Service) GetForCreate(obj interface{}) ([]runtime.Object, error) {
	ros, err := s.newRuntimeObjects(obj)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	return ros, nil
}

// GetForDelete returns the Kubernetes runtime object for the master resource
// being used in reconciliation loops on delete events. Note that deleting a
// master resources happens implicity by deleting the namespace the master
// resource is attached to. This is why we do not return any implementation
// here, but nil. That causes the reconcilliation to ignore the deletion for
// this resource.
func (s *Service) GetForDelete(obj interface{}) ([]runtime.Object, error) {
	return nil, nil
}

func (s *Service) newRuntimeObjects(obj interface{}) ([]runtime.Object, error) {
	var err error
	var runtimeObjects []runtime.Object

	var initContainers []apiv1.Container
	{
		initContainers, err = s.newInitContainers(obj)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var flannelContainers []apiv1.Container
	{
		flannelContainers, err = s.flannel.Containers(obj)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var flannelInitContainers []apiv1.Container
	{
		flannelInitContainers, err = s.flannel.InitContainers(obj)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var flannelVolumes []apiv1.Volume
	{
		flannelVolumes, err = s.flannel.Volumes(obj)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var podAffinity *apiv1.Affinity
	{
		podAffinity, err = s.newPodAfinity(obj)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var newDeployments []*extensionsv1.Deployment
	{
		newDeployments, err = s.newDeployments(obj)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
		for i, _ := range newDeployments {
			newDeployments[i].Spec.Template.Spec.Affinity = podAffinity
			newDeployments[i].Spec.Template.Spec.InitContainers = append(flannelInitContainers, initContainers...)
			newDeployments[i].Spec.Template.Spec.Containers = append(flannelContainers, newDeployments[i].Spec.Template.Spec.Containers...)
			newDeployments[i].Spec.Template.Spec.Volumes = append(flannelVolumes, newDeployments[i].Spec.Template.Spec.Volumes...)
		}
	}

	var newIngresses []*extensionsv1.Ingress
	{
		newIngresses, err = s.newIngresses(obj)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var newService *apiv1.Service
	{
		newService, err = s.newService(obj)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	for _, i := range newIngresses {
		runtimeObjects = append(runtimeObjects, i)
	}
	runtimeObjects = append(runtimeObjects, newService)
	for _, d := range newDeployments {
		runtimeObjects = append(runtimeObjects, d)
	}

	return runtimeObjects, nil
}
