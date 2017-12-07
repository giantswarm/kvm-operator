package service

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/client/k8sclient"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/framework/resource/retryresource"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/operatorkit/tpr"
	"github.com/giantswarm/randomkeytpr"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/service/cloudconfigv1"
	"github.com/giantswarm/kvm-operator/service/keyv1"
	"github.com/giantswarm/kvm-operator/service/messagecontext"
	"github.com/giantswarm/kvm-operator/service/resource/configmapv1"
	"github.com/giantswarm/kvm-operator/service/resource/deploymentv1"
	"github.com/giantswarm/kvm-operator/service/resource/ingressv1"
	"github.com/giantswarm/kvm-operator/service/resource/namespacev1"
	"github.com/giantswarm/kvm-operator/service/resource/podv1"
	"github.com/giantswarm/kvm-operator/service/resource/pvcv1"
	"github.com/giantswarm/kvm-operator/service/resource/servicev1"
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
		c := k8sclient.DefaultConfig()

		c.Address = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
		c.Logger = config.Logger
		c.InCluster = config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster)
		c.TLS.CAFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile)
		c.TLS.CrtFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile)
		c.TLS.KeyFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile)

		k8sClient, err = k8sclient.New(c)
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

	var keyWatcher randomkeytpr.Searcher
	{
		keyConfig := randomkeytpr.DefaultServiceConfig()
		keyConfig.K8sClient = k8sClient
		keyConfig.Logger = config.Logger
		keyWatcher, err = randomkeytpr.NewService(keyConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ccService *cloudconfigv1.CloudConfig
	{
		c := cloudconfigv1.DefaultConfig()

		c.Logger = config.Logger

		ccService, err = cloudconfigv1.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configMapResource framework.Resource
	{
		c := configmapv1.DefaultConfig()

		c.CertWatcher = certWatcher
		c.CloudConfig = ccService
		c.K8sClient = k8sClient
		c.KeyWatcher = keyWatcher
		c.Logger = config.Logger

		configMapResource, err = configmapv1.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var deploymentResource framework.Resource
	{
		c := deploymentv1.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		deploymentResource, err = deploymentv1.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ingressResource framework.Resource
	{
		c := ingressv1.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		ingressResource, err = ingressv1.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var namespaceResource framework.Resource
	{
		c := namespacev1.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		namespaceResource, err = namespacev1.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var pvcResource framework.Resource
	{
		c := pvcv1.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		pvcResource, err = pvcv1.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var serviceResource framework.Resource
	{
		c := servicev1.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		serviceResource, err = servicev1.New(c)
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
		c := informer.DefaultConfig()

		c.WatcherFactory = newWatcherFactory

		c.RateWait = 10 * time.Second
		c.ResyncPeriod = 5 * time.Minute

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		message := messagecontext.NewMessage()
		ctx = messagecontext.NewContext(ctx, message)

		return ctx, nil
	}

	var customObjectFramework *framework.Framework
	{
		c := framework.DefaultConfig()

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
		c := k8sclient.DefaultConfig()

		c.Address = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
		c.Logger = config.Logger
		c.InCluster = config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster)
		c.TLS.CAFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile)
		c.TLS.CrtFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile)
		c.TLS.KeyFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile)

		k8sClient, err = k8sclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var podResource framework.Resource
	{
		c := podv1.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		podResource, err = podv1.New(c)
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

	var newWatcherFactory informer.WatcherFactory
	{
		newWatcherFactory = func() (watch.Interface, error) {
			options := apismetav1.ListOptions{
				LabelSelector: fmt.Sprintf("%s=%s", keyv1.PodWatcherLabel, config.Name),
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
		c := informer.DefaultConfig()

		c.WatcherFactory = newWatcherFactory

		c.RateWait = 10 * time.Second
		c.ResyncPeriod = 5 * time.Minute

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
