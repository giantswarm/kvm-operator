package service

import (
	"context"
	"fmt"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/context/updateallowedcontext"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/framework/resource/retryresource"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/randomkeys"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/kvm-operator/service/cloudconfigv2"
	"github.com/giantswarm/kvm-operator/service/cloudconfigv3"
	"github.com/giantswarm/kvm-operator/service/keyv3"
	"github.com/giantswarm/kvm-operator/service/resource/configmapv2"
	"github.com/giantswarm/kvm-operator/service/resource/configmapv3"
	"github.com/giantswarm/kvm-operator/service/resource/deploymentv2"
	"github.com/giantswarm/kvm-operator/service/resource/deploymentv3"
	"github.com/giantswarm/kvm-operator/service/resource/ingressv2"
	"github.com/giantswarm/kvm-operator/service/resource/namespacev2"
	"github.com/giantswarm/kvm-operator/service/resource/podv2"
	"github.com/giantswarm/kvm-operator/service/resource/pvcv2"
	"github.com/giantswarm/kvm-operator/service/resource/servicev2"
)

const (
	ResourceRetries uint64 = 3
)

func newCRDFramework(config Config) (*framework.Framework, error) {
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Flag must not be empty")
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Viper must not be empty")
	}

	var err error

	var restConfig *rest.Config
	{
		c := k8srestconfig.DefaultConfig()

		c.Logger = config.Logger

		c.Address = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
		c.InCluster = config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster)
		c.TLS.CAFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile)
		c.TLS.CrtFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile)
		c.TLS.KeyFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile)

		restConfig, err = k8srestconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	clientSet, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sExtClient, err := apiextensionsclient.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var crdClient *k8scrdclient.CRDClient
	{
		c := k8scrdclient.DefaultConfig()

		c.K8sExtClient = k8sExtClient
		c.Logger = config.Logger

		crdClient, err = k8scrdclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var certSearcher certs.Interface
	{
		c := certs.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		certSearcher, err = certs.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
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

	var ccServiceV3 *cloudconfigv3.CloudConfig
	{
		c := cloudconfigv3.DefaultConfig()

		c.Logger = config.Logger

		ccServiceV3, err = cloudconfigv3.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var keyWatcher randomkeys.Interface
	{
		keyConfig := randomkeys.DefaultConfig()
		keyConfig.K8sClient = k8sClient
		keyConfig.Logger = config.Logger
		keyWatcher, err = randomkeys.NewSearcher(keyConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configMapResourceV2 framework.Resource
	{
		c := configmapv2.DefaultConfig()

		c.CertSearcher = certSearcher
		c.CloudConfig = ccServiceV2
		c.K8sClient = k8sClient
		c.Logger = config.Logger

		configMapResourceV2, err = configmapv2.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configMapResourceV3 framework.Resource
	{
		c := configmapv3.DefaultConfig()

		c.CertSearcher = certSearcher
		c.CloudConfig = ccServiceV3
		c.K8sClient = k8sClient
		c.KeyWatcher = keyWatcher
		c.Logger = config.Logger

		configMapResourceV3, err = configmapv3.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	var deploymentResourceV2 framework.Resource
	{
		c := deploymentv2.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		deploymentResourceV2, err = deploymentv2.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	var deploymentResourceV3 framework.Resource
	{
		c := deploymentv3.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		deploymentResourceV3, err = deploymentv3.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ingressResource framework.Resource
	{
		c := ingressv2.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		ingressResource, err = ingressv2.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var namespaceResource framework.Resource
	{
		c := namespacev2.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		namespaceResource, err = namespacev2.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var pvcResource framework.Resource
	{
		c := pvcv2.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		pvcResource, err = pvcv2.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var serviceResource framework.Resource
	{
		c := servicev2.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		serviceResource, err = servicev2.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourcesV2 []framework.Resource
	{
		resourcesV2 = []framework.Resource{
			namespaceResource,

			configMapResourceV2,
			deploymentResourceV2,
			ingressResource,
			pvcResource,
			serviceResource,
		}

		retryWrapConfig := retryresource.DefaultWrapConfig()

		retryWrapConfig.BackOffFactory = func() backoff.BackOff { return backoff.WithMaxTries(backoff.NewExponentialBackOff(), ResourceRetries) }
		retryWrapConfig.Logger = config.Logger

		resourcesV2, err = retryresource.Wrap(resourcesV2, retryWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		metricsWrapConfig := metricsresource.DefaultWrapConfig()

		metricsWrapConfig.Name = config.Name

		resourcesV2, err = metricsresource.Wrap(resourcesV2, metricsWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourcesV3 []framework.Resource
	{
		resourcesV3 = []framework.Resource{
			namespaceResource,

			configMapResourceV3,
			deploymentResourceV3,
			ingressResource,
			pvcResource,
			serviceResource,
		}

		retryWrapConfig := retryresource.DefaultWrapConfig()

		retryWrapConfig.BackOffFactory = func() backoff.BackOff { return backoff.WithMaxTries(backoff.NewExponentialBackOff(), ResourceRetries) }
		retryWrapConfig.Logger = config.Logger

		resourcesV3, err = retryresource.Wrap(resourcesV3, retryWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		metricsWrapConfig := metricsresource.DefaultWrapConfig()

		metricsWrapConfig.Name = config.Name

		resourcesV3, err = metricsresource.Wrap(resourcesV3, metricsWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// We provide a map of resource lists keyed by the version bundle version
	// to the resource router.
	versionedResources := map[string][]framework.Resource{
		"1.1.0": resourcesV3,
		"1.0.0": resourcesV2,
		"0.1.0": resourcesV2,
		"":      resourcesV2,
	}

	var newInformer *informer.Informer
	{
		c := informer.DefaultConfig()

		c.Watcher = clientSet.ProviderV1alpha1().KVMConfigs("")

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		{
			customObject, err := keyv3.ToCustomObject(obj)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			updateEnabled := config.Viper.GetBool(config.Flag.Service.Guest.Update.Enabled)
			versionBundleVersion := keyv3.VersionBundleVersion(customObject)

			if updateEnabled && versionBundleVersion >= "1.1.0" {
				updateallowedcontext.SetUpdateAllowed(ctx)
			}
		}

		return ctx, nil
	}

	var crdFramework *framework.Framework
	{
		c := framework.DefaultConfig()

		c.CRD = v1alpha1.NewKVMConfigCRD()
		c.CRDClient = crdClient
		c.Informer = newInformer
		c.InitCtxFunc = initCtxFunc
		c.Logger = config.Logger
		c.ResourceRouter = newResourceRouter(versionedResources)

		crdFramework, err = framework.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return crdFramework, nil
}

func newPodFramework(config Config) (*framework.Framework, error) {
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Flag must not be empty")
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Viper must not be empty")
	}

	var err error

	var restConfig *rest.Config
	{
		c := k8srestconfig.DefaultConfig()

		c.Logger = config.Logger

		c.Address = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
		c.InCluster = config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster)
		c.TLS.CAFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile)
		c.TLS.CrtFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile)
		c.TLS.KeyFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile)

		restConfig, err = k8srestconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var podResource framework.Resource
	{
		c := podv2.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		podResource, err = podv2.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resources []framework.Resource
	{
		resources = []framework.Resource{
			podResource,
		}

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

	var newInformer *informer.Informer
	{
		c := informer.DefaultConfig()

		c.Watcher = k8sClient.CoreV1().Pods("")

		c.ListOptions = apismetav1.ListOptions{
			LabelSelector: fmt.Sprintf("pod-watcher=%s", keyv3.PodWatcherLabel),
		}
		fmt.Printf("c.ListOptions: %#v\n", c.ListOptions)

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		return ctx, nil
	}

	var podFramework *framework.Framework
	{
		c := framework.DefaultConfig()

		c.Informer = newInformer
		c.InitCtxFunc = initCtxFunc
		c.Logger = config.Logger
		c.ResourceRouter = framework.DefaultResourceRouter(resources)

		podFramework, err = framework.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return podFramework, nil
}
