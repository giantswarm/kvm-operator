package cloudconfigv2

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certificatetpr"
	cloudconfig "github.com/giantswarm/k8scloudconfig"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_1_1_0"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/randomkeytpr"
)

const (
	FileOwner      = "root:root"
	FilePermission = 0700
)

// Config represents the configuration used to create a cloud config service.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new cloud config
// service by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger: nil,
	}
}

// CloudConfig implements the cloud config service interface.
type CloudConfig struct {
	// Dependencies.
	logger micrologger.Logger
}

// New creates a new configured cloud config service.
func New(config Config) (*CloudConfig, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	newCloudConfig := &CloudConfig{
		// Dependencies.
		logger: config.Logger,
	}

	return newCloudConfig, nil
}

// NewMasterTemplate generates a new worker cloud config template and returns it
// as a base64 encoded string.
func (c *CloudConfig) NewMasterTemplate(customObject v1alpha1.KVMConfig, certs certificatetpr.AssetsBundle, node v1alpha1.ClusterNode, keys map[randomkeytpr.Key][]byte) (string, error) {
	var err error
	var template string
	switch customObject.Spec.Cluster.Version {
	case string(cloudconfig.V_1_1_0):
		template, err = v1_1_0MasterTemplate(customObject, certs, node)
		if err != nil {
			return "", microerror.Mask(err)
		}
	case string(cloudconfig.V_2_0_0):
		template, err = v2_0_0MasterTemplate(customObject, certs, node, keys)
		if err != nil {
			return "", microerror.Mask(err)
		}

	default:
		return "", microerror.Maskf(notFoundError, "k8scloudconfig version '%s'", customObject.Spec.Cluster.Version)
	}

	return template, nil
}

// NewWorkerTemplate generates a new worker cloud config template and returns it
// as a base64 encoded string.
func (c *CloudConfig) NewWorkerTemplate(customObject v1alpha1.KVMConfig, certs certificatetpr.AssetsBundle, node v1alpha1.ClusterNode) (string, error) {
	var err error

	var params k8scloudconfig.Params
	{
		params.Cluster = customObject.Spec.Cluster
		params.Extension = &workerExtension{
			certs: certs,
		}
		params.Node = node
	}

	var newCloudConfig *k8scloudconfig.CloudConfig
	{
		newCloudConfig, err = k8scloudconfig.NewCloudConfig(k8scloudconfig.WorkerTemplate, params)
		if err != nil {
			return "", microerror.Mask(err)
		}

		err = newCloudConfig.ExecuteTemplate()
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	return newCloudConfig.Base64(), nil
}
