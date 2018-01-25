package service

import (
	"github.com/cenkalti/backoff"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/kvm-operator/service/cloudconfigv2"
	"github.com/giantswarm/kvm-operator/service/resource/configmapv2"
	"github.com/giantswarm/kvm-operator/service/resource/deploymentv2"
	"github.com/giantswarm/kvm-operator/service/resource/ingressv2"
	"github.com/giantswarm/kvm-operator/service/resource/namespacev2"
	"github.com/giantswarm/kvm-operator/service/resource/pvcv2"
	"github.com/giantswarm/kvm-operator/service/resource/serviceaccountv2"
	"github.com/giantswarm/kvm-operator/service/resource/servicev2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/framework/resource/retryresource"
	"github.com/giantswarm/randomkeys"
	"k8s.io/client-go/kubernetes"
)

type kvmConfigResourcesV2Config struct {
	// Dependencies.

	CertsSearcher      certs.Interface
	K8sClient          kubernetes.Interface
	Logger             micrologger.Logger
	RandomkeysSearcher randomkeys.Interface

	// Settings.

	// Name is the project name.
	Name string
}

func defaultKVMConfigResourcesV2Config() kvmConfigResourcesV2Config {
	return kvmConfigResourcesV2Config{
		CertsSearcher:      nil,
		K8sClient:          nil,
		Logger:             nil,
		RandomkeysSearcher: nil,

		Name: "",
	}
}

func newKVMConfigResourcesV2(config kvmConfigResourcesV2Config) ([]framework.Resource, error) {
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

	var ccServiceV2 *cloudconfigv2.CloudConfig
	{
		c := cloudconfigv2.DefaultConfig()

		c.Logger = config.Logger

		ccServiceV2, err = cloudconfigv2.New(c)
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

	var serviceAccountResourceV2 framework.Resource
	{
		c := serviceaccountv2.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		serviceAccountResourceV2, err = serviceaccountv2.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configMapResourceV2 framework.Resource
	{
		c := configmapv2.DefaultConfig()

		c.CertSearcher = config.CertsSearcher
		c.CloudConfig = ccServiceV2
		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		configMapResourceV2, err = configmapv2.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var deploymentResourceV2 framework.Resource
	{
		c := deploymentv2.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		deploymentResourceV2, err = deploymentv2.New(c)
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
		serviceAccountResourceV2,
		configMapResourceV2,
		deploymentResourceV2,
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
