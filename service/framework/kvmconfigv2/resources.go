package kvmconfigv2

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

const (
	ResourceRetries uint64 = 3
)

type ResourcesConfig struct {
	// Dependencies.

	CertsSearcher      certs.Interface
	K8sClient          kubernetes.Interface
	Logger             micrologger.Logger
	RandomkeysSearcher randomkeys.Interface

	// Settings.

	// Name is the project name.
	Name string
}

func NewResources(config ResourcesConfig) ([]framework.Resource, error) {
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

	var cloudConfig *cloudconfigv2.CloudConfig
	{
		c := cloudconfigv2.DefaultConfig()

		c.Logger = config.Logger

		cloudConfig, err = cloudconfigv2.New(c)
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

	var serviceAccountResource framework.Resource
	{
		c := serviceaccountv2.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		serviceAccountResource, err = serviceaccountv2.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configMapResource framework.Resource
	{
		c := configmapv2.DefaultConfig()

		c.CertSearcher = config.CertsSearcher
		c.CloudConfig = cloudConfig
		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		configMapResource, err = configmapv2.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var deploymentResource framework.Resource
	{
		c := deploymentv2.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		deploymentResource, err = deploymentv2.New(c)
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
		serviceAccountResource,
		configMapResource,
		deploymentResource,
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
