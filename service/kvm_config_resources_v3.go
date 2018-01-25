package service

import (
	"github.com/cenkalti/backoff"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/kvm-operator/service/cloudconfigv3"
	"github.com/giantswarm/kvm-operator/service/resource/configmapv3"
	"github.com/giantswarm/kvm-operator/service/resource/deploymentv3"
	"github.com/giantswarm/kvm-operator/service/resource/ingressv2"
	"github.com/giantswarm/kvm-operator/service/resource/namespacev2"
	"github.com/giantswarm/kvm-operator/service/resource/pvcv2"
	"github.com/giantswarm/kvm-operator/service/resource/serviceaccountv3"
	"github.com/giantswarm/kvm-operator/service/resource/servicev2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/framework/resource/retryresource"
	"github.com/giantswarm/randomkeys"
	"k8s.io/client-go/kubernetes"
)

type kvmConfigResourcesV3Config struct {
	// Dependencies.

	CertsSearcher      certs.Interface
	K8sClient          kubernetes.Interface
	Logger             micrologger.Logger
	RandomkeysSearcher randomkeys.Interface

	// Settings.

	// Name is the project name.
	Name string
}

func defaultKVMConfigResourcesV3Config() kvmConfigResourcesV3Config {
	return kvmConfigResourcesV3Config{
		CertsSearcher:      nil,
		K8sClient:          nil,
		Logger:             nil,
		RandomkeysSearcher: nil,

		Name: "",
	}
}

func newKVMConfigResourcesV3(config kvmConfigResourcesV3Config) ([]framework.Resource, error) {
	var err error

	if config.CertsSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CertsSearcher must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.RandomkeysSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.RandomkeysSearcher must not be empty")
	}

	if config.Name == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Name must not be empty")
	}

	var ccServiceV3 *cloudconfigv3.CloudConfig
	{
		c := cloudconfigv3.DefaultConfig()

		c.Logger = config.Logger

		ccServiceV3, err = cloudconfigv3.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var namespaceResource framework.Resource
	{
		c := namespacev2.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		namespaceResource, err = namespacev2.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var serviceAccountResourceV3 framework.Resource
	{
		c := serviceaccountv3.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		serviceAccountResourceV3, err = serviceaccountv3.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configMapResourceV3 framework.Resource
	{
		c := configmapv3.DefaultConfig()

		c.CertSearcher = config.CertsSearcher
		c.CloudConfig = ccServiceV3
		c.K8sClient = config.K8sClient
		c.KeyWatcher = config.RandomkeysSearcher
		c.Logger = config.Logger

		configMapResourceV3, err = configmapv3.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var deploymentResourceV3 framework.Resource
	{
		c := deploymentv3.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		deploymentResourceV3, err = deploymentv3.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ingressResource framework.Resource
	{
		c := ingressv2.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		ingressResource, err = ingressv2.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var pvcResource framework.Resource
	{
		c := pvcv2.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		pvcResource, err = pvcv2.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var serviceResource framework.Resource
	{
		c := servicev2.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		serviceResource, err = servicev2.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []framework.Resource{
		namespaceResource,
		serviceAccountResourceV3,
		configMapResourceV3,
		deploymentResourceV3,
		ingressResource,
		pvcResource,
		serviceResource,
	}

	// Wrap resources with retry and metrics.
	{
		retryWrapConfig := retryresource.DefaultWrapConfig()

		retryWrapConfig.BackOffFactory = func() backoff.BackOff { return backoff.WithMaxTries(backoff.NewExponentialBackOff(), ResourceRetries) }
		retryWrapConfig.Logger = config.Logger

		resources, err = retryresource.Wrap(resources, retryWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		metricsWrapConfig := metricsresource.DefaultWrapConfig()

		metricsWrapConfig.Name = config.Name

		resources, err = metricsresource.Wrap(resources, metricsWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resources, nil
}
