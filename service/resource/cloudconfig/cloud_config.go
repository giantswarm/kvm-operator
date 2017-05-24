package cloudconfig

import (
	cloudconfig "github.com/giantswarm/k8scloudconfig"
	"github.com/giantswarm/kvmtpr"
	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	certkit "github.com/giantswarm/operatorkit/secret/cert"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"

	"github.com/giantswarm/kvm-operator/service/resource"
)

const (
	FileOwner      = "root:root"
	FilePermission = 0700
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	CertWatcher *certkit.Service
	Logger      micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new service by
// best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		CertWatcher: nil,
		Logger:      nil,
	}
}

// New creates a new configured service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.CertWatcher == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "config.CertWatcher must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "config.Logger must not be empty")
	}

	newService := &Service{
		// Dependencies.
		certWatcher: config.CertWatcher,
		logger:      config.Logger,
	}

	return newService, nil
}

// Service implements the cloud config service.
type Service struct {
	// Dependencies.
	certWatcher *certkit.Service
	logger      micrologger.Logger
}

// GetForCreate returns the Kubernetes runtime object for the cloud config
// resource being used in reconciliation loops on create events.
func (s *Service) GetForCreate(obj interface{}) ([]runtime.Object, error) {
	ros, err := s.newRuntimeObjects(obj)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	return ros, nil
}

// GetForDelete returns the Kubernetes runtime object for the cloud config
// resource being used in reconciliation loops on delete events. Note that
// deleting a cloud config resources happens implicity by deleting the namespace
// the cloud config resource is attached to. This is why we do not return any
// implementation here, but nil. That causes the reconcilliation to ignore the
// deletion for this resource.
func (s *Service) GetForDelete(obj interface{}) ([]runtime.Object, error) {
	return nil, nil
}

// newConfigMap creates a new Kubernetes configmap using the provided
// information. customObject is used for name and label creation. params serves
// as structure being injected into the template execution to interpolate
// variables. prefix can be either "master" or "worker" and is used to prefix
// the configmap name.
func (s *Service) newConfigMap(customObject kvmtpr.CustomObject, template string, params cloudconfig.Params, prefix string) (*v1.ConfigMap, error) {
	var err error

	var newCloudConfig *cloudconfig.CloudConfig
	{
		newCloudConfig, err = cloudconfig.NewCloudConfig(template, params)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}

		err = newCloudConfig.ExecuteTemplate()
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var newConfigMap *v1.ConfigMap
	{
		newConfigMap = &v1.ConfigMap{
			ObjectMeta: v1.ObjectMeta{
				Name: resource.ConfigMapName(customObject, params.Node, prefix),
				Labels: map[string]string{
					"cluster":  resource.ClusterID(customObject),
					"customer": resource.ClusterCustomer(customObject),
				},
			},
			Data: map[string]string{
				"user_data": newCloudConfig.String(),
			},
		}
	}

	return newConfigMap, nil
}

func (s *Service) newRuntimeObjects(obj interface{}) ([]runtime.Object, error) {
	var err error
	var runtimeObjects []runtime.Object

	customObject, ok := obj.(*kvmtpr.CustomObject)
	if !ok {
		return nil, microerror.MaskAnyf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, obj)
	}

	certs, err := s.certWatcher.SearchCerts(customObject.Spec.Cluster.Cluster.ID)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	for _, mn := range customObject.Spec.Cluster.Masters {
		newExtension := &MasterExtension{
			certs: certs,
		}

		var params cloudconfig.Params
		{
			params.Cluster = customObject.Spec.Cluster
			params.Extension = newExtension
			params.Node = mn
		}

		cm, err := s.newConfigMap(*customObject, cloudconfig.MasterTemplate, params, "master")
		if err != nil {
			return nil, microerror.MaskAny(err)
		}

		runtimeObjects = append(runtimeObjects, cm)
	}

	for _, wn := range customObject.Spec.Cluster.Workers {
		newExtension := &WorkerExtension{
			certs: certs,
		}

		var params cloudconfig.Params
		{
			params.Cluster = customObject.Spec.Cluster
			params.Extension = newExtension
			params.Node = wn
		}

		cm, err := s.newConfigMap(*customObject, cloudconfig.WorkerTemplate, params, "worker")
		if err != nil {
			return nil, microerror.MaskAny(err)
		}

		runtimeObjects = append(runtimeObjects, cm)
	}

	return runtimeObjects, nil
}
