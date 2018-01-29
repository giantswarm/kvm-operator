package kvmconfigv3

import (
	"github.com/cenkalti/backoff"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/kvm-operator/service/cloudconfigv3"
	"github.com/giantswarm/kvm-operator/service/resource/clusterrolebindingv3"
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

const (
	ResourceRetries uint64 = 3
)

type ResourcesConfig struct {
	CertsSearcher      certs.Interface
	K8sClient          kubernetes.Interface
	Logger             micrologger.Logger
	RandomkeysSearcher randomkeys.Interface

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

	var cloudConfig *cloudconfigv3.CloudConfig
	{
		c := cloudconfigv3.DefaultConfig()

		c.Logger = config.Logger

		cloudConfig, err = cloudconfigv3.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterRoleBindingV3 framework.Resource
	{
		c := clusterrolebindingv3.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		clusterRoleBindingV3, err = clusterrolebindingv3.New(c)
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
		c := serviceaccountv3.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		serviceAccountResource, err = serviceaccountv3.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configMapResource framework.Resource
	{
		c := configmapv3.DefaultConfig()

		c.CertSearcher = config.CertsSearcher
		c.CloudConfig = cloudConfig
		c.K8sClient = config.K8sClient
		c.KeyWatcher = config.RandomkeysSearcher
		c.Logger = config.Logger

		configMapResource, err = configmapv3.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var deploymentResource framework.Resource
	{
		c := deploymentv3.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		deploymentResource, err = deploymentv3.New(c)
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
		clusterRoleBindingV3,
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
