package service

import (
	"context"
	"time"

	"github.com/cenk/backoff"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/client/k8sclient"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/resource/logresource"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/framework/resource/retryresource"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/operatorkit/tpr"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/service/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/key"
	"github.com/giantswarm/kvm-operator/service/messagecontext"
	configmapresource "github.com/giantswarm/kvm-operator/service/resource/configmap"
	deploymentresource "github.com/giantswarm/kvm-operator/service/resource/deployment"
	ingressresource "github.com/giantswarm/kvm-operator/service/resource/ingress"
	namespaceresource "github.com/giantswarm/kvm-operator/service/resource/namespace"
	podresource "github.com/giantswarm/kvm-operator/service/resource/pod"
	pvcresource "github.com/giantswarm/kvm-operator/service/resource/pvc"
	serviceresource "github.com/giantswarm/kvm-operator/service/resource/service"
)

const (
	ResourceRetries uint64 = 3
)

func newCustomObjectFramework(config Config) (*framework.Framework, error) {
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Flag must not be empty")
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Viper must not be empty")
	}

	var err error

	var k8sClient kubernetes.Interface
	{
		k8sConfig := k8sclient.DefaultConfig()

		k8sConfig.Address = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
		k8sConfig.Logger = config.Logger
		k8sConfig.InCluster = config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster)
		k8sConfig.TLS.CAFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile)
		k8sConfig.TLS.CrtFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile)
		k8sConfig.TLS.KeyFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile)

		k8sClient, err = k8sclient.New(k8sConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var certWatcher certificatetpr.Searcher
	{
		certConfig := certificatetpr.DefaultServiceConfig()

		certConfig.K8sClient = k8sClient
		certConfig.Logger = config.Logger

		certWatcher, err = certificatetpr.NewService(certConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ccService *cloudconfig.CloudConfig
	{
		ccServiceConfig := cloudconfig.DefaultConfig()

		ccServiceConfig.Logger = config.Logger

		ccService, err = cloudconfig.New(ccServiceConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configMapResource framework.Resource
	{
		configMapConfig := configmapresource.DefaultConfig()

		configMapConfig.CertWatcher = certWatcher
		configMapConfig.CloudConfig = ccService
		configMapConfig.K8sClient = k8sClient
		configMapConfig.Logger = config.Logger

		configMapResource, err = configmapresource.New(configMapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var deploymentResource framework.Resource
	{
		deploymentConfig := deploymentresource.DefaultConfig()

		deploymentConfig.K8sClient = k8sClient
		deploymentConfig.Logger = config.Logger

		deploymentResource, err = deploymentresource.New(deploymentConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ingressResource framework.Resource
	{
		ingressConfig := ingressresource.DefaultConfig()

		ingressConfig.K8sClient = k8sClient
		ingressConfig.Logger = config.Logger

		ingressResource, err = ingressresource.New(ingressConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var namespaceResource framework.Resource
	{
		namespaceConfig := namespaceresource.DefaultConfig()

		namespaceConfig.K8sClient = k8sClient
		namespaceConfig.Logger = config.Logger

		namespaceResource, err = namespaceresource.New(namespaceConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var pvcResource framework.Resource
	{
		pvcConfig := pvcresource.DefaultConfig()

		pvcConfig.K8sClient = k8sClient
		pvcConfig.Logger = config.Logger

		pvcResource, err = pvcresource.New(pvcConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var serviceResource framework.Resource
	{
		serviceConfig := serviceresource.DefaultConfig()

		serviceConfig.K8sClient = k8sClient
		serviceConfig.Logger = config.Logger

		serviceResource, err = serviceresource.New(serviceConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resources []framework.Resource
	{
		resources = []framework.Resource{
			namespaceResource,

			configMapResource,
			deploymentResource,
			ingressResource,
			pvcResource,
			serviceResource,
		}

		logWrapConfig := logresource.DefaultWrapConfig()

		logWrapConfig.Logger = config.Logger

		resources, err = logresource.Wrap(resources, logWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
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

	var newTPR *tpr.TPR
	{
		c := tpr.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		c.Description = kvmtpr.Description
		c.Name = kvmtpr.Name
		c.Version = kvmtpr.VersionV1

		newTPR, err = tpr.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var newWatcherFactory informer.WatcherFactory
	{
		zeroObjectFactory := &informer.ZeroObjectFactoryFuncs{
			NewObjectFunc:     func() runtime.Object { return &kvmtpr.CustomObject{} },
			NewObjectListFunc: func() runtime.Object { return &kvmtpr.List{} },
		}
		newWatcherFactory = informer.NewWatcherFactory(k8sClient.Discovery().RESTClient(), newTPR.WatchEndpoint(""), zeroObjectFactory)
	}

	var newInformer *informer.Informer
	{
		informerConfig := informer.DefaultConfig()

		informerConfig.WatcherFactory = newWatcherFactory

		informerConfig.RateWait = 10 * time.Second
		informerConfig.ResyncPeriod = 5 * time.Minute

		newInformer, err = informer.New(informerConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		ctx = messagecontext.NewContext(ctx, messagecontext.NewMessage())

		return ctx, nil
	}

	var customObjectFramework *framework.Framework
	{
		c := framework.DefaultConfig()

		c.BackOffFactory = framework.DefaultBackOffFactory()
		c.Informer = newInformer
		c.InitCtxFunc = initCtxFunc
		c.Logger = config.Logger
		c.ResourceRouter = framework.DefaultResourceRouter(resources)
		c.TPR = newTPR

		customObjectFramework, err = framework.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return customObjectFramework, nil
}

func newPodFramework(config Config) (*framework.Framework, error) {
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Flag must not be empty")
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Viper must not be empty")
	}

	var err error

	var k8sClient kubernetes.Interface
	{
		k8sConfig := k8sclient.DefaultConfig()

		k8sConfig.Address = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
		k8sConfig.Logger = config.Logger
		k8sConfig.InCluster = config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster)
		k8sConfig.TLS.CAFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile)
		k8sConfig.TLS.CrtFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile)
		k8sConfig.TLS.KeyFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile)

		k8sClient, err = k8sclient.New(k8sConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var podResource framework.Resource
	{
		c := podresource.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		podResource, err = podresource.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resources []framework.Resource
	{
		resources = []framework.Resource{
			podResource,
		}

		logWrapConfig := logresource.DefaultWrapConfig()

		logWrapConfig.Logger = config.Logger

		resources, err = logresource.Wrap(resources, logWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
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

	var newWatcherFactory informer.WatcherFactory
	{
		newWatcherFactory = func() (watch.Interface, error) {
			options := apismetav1.ListOptions{
				LabelSelector: key.PodWatcherLabel,
			}

			watcher, err := k8sClient.CoreV1().Pods("").Watch(options)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			return watcher, nil
		}
	}

	var newInformer *informer.Informer
	{
		informerConfig := informer.DefaultConfig()

		informerConfig.WatcherFactory = newWatcherFactory

		informerConfig.RateWait = 10 * time.Second
		informerConfig.ResyncPeriod = 5 * time.Minute

		newInformer, err = informer.New(informerConfig)
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

		c.BackOffFactory = framework.DefaultBackOffFactory()
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
