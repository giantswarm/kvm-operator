package v3

import (
	"context"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/context/updateallowedcontext"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/framework/resource/retryresource"
	"github.com/giantswarm/randomkeys"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/service/kvmconfig/v3/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v3/key"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v3/resource/clusterrolebinding"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v3/resource/configmap"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v3/resource/deployment"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v3/resource/ingress"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v3/resource/namespace"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v3/resource/pvc"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v3/resource/service"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v3/resource/serviceaccount"
)

const (
	ResourceRetries uint64 = 3
)

type ResourceSetConfig struct {
	CertsSearcher      certs.Interface
	K8sClient          kubernetes.Interface
	Logger             micrologger.Logger
	RandomkeysSearcher randomkeys.Interface

	GuestUpdateEnabled    bool
	HandledVersionBundles []string
	// Name is the project name.
	Name string
}

func NewResourceSet(config ResourceSetConfig) (*framework.ResourceSet, error) {
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

	if config.HandledVersionBundles == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.HandledVersionBundles must not be empty")
	}
	if config.Name == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Name must not be empty")
	}

	var cloudConfig *cloudconfig.CloudConfig
	{
		c := cloudconfig.DefaultConfig()

		c.Logger = config.Logger

		cloudConfig, err = cloudconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var namespaceResource framework.Resource
	{
		c := namespace.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		namespaceResource, err = namespace.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var serviceAccountResource framework.Resource
	{
		c := serviceaccount.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		serviceAccountResource, err = serviceaccount.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterRoleBinding framework.Resource
	{
		c := clusterrolebinding.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		clusterRoleBinding, err = clusterrolebinding.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configMapResource framework.Resource
	{
		c := configmap.DefaultConfig()

		c.CertSearcher = config.CertsSearcher
		c.CloudConfig = cloudConfig
		c.K8sClient = config.K8sClient
		c.KeyWatcher = config.RandomkeysSearcher
		c.Logger = config.Logger

		configMapResource, err = configmap.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var deploymentResource framework.Resource
	{
		c := deployment.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		deploymentResource, err = deployment.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ingressResource framework.Resource
	{
		c := ingress.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		ingressResource, err = ingress.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var pvcResource framework.Resource
	{
		c := pvc.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		pvcResource, err = pvc.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var serviceResource framework.Resource
	{
		c := service.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		serviceResource, err = service.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []framework.Resource{
		namespaceResource,
		serviceAccountResource,
		clusterRoleBinding,
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

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		if config.GuestUpdateEnabled {
			updateallowedcontext.SetUpdateAllowed(ctx)
		}

		return ctx, nil
	}

	handlesFunc := func(obj interface{}) bool {
		kvmConfig, err := key.ToCustomObject(obj)
		if err != nil {
			return false
		}
		versionBundleVersion := key.VersionBundleVersion(kvmConfig)

		for _, v := range config.HandledVersionBundles {
			if versionBundleVersion == v {
				return true
			}
		}

		return false
	}

	var resourceSet *framework.ResourceSet
	{
		c := framework.ResourceSetConfig{

			Handles:   handlesFunc,
			InitCtx:   initCtxFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		resourceSet, err = framework.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceSet, nil
}
