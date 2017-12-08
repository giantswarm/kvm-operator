package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/clustertpr/spec/kubernetes/ssh"
	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/kvmtpr/spec/kvm"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/framework/resource/retryresource"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/operatorkit/tpr"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/kvm-operator/service/cloudconfigv1"
	"github.com/giantswarm/kvm-operator/service/cloudconfigv2"
	"github.com/giantswarm/kvm-operator/service/keyv1"
	"github.com/giantswarm/kvm-operator/service/messagecontext"
	"github.com/giantswarm/kvm-operator/service/resource/configmapv1"
	"github.com/giantswarm/kvm-operator/service/resource/configmapv2"
	"github.com/giantswarm/kvm-operator/service/resource/deploymentv1"
	"github.com/giantswarm/kvm-operator/service/resource/deploymentv2"
	"github.com/giantswarm/kvm-operator/service/resource/ingressv1"
	"github.com/giantswarm/kvm-operator/service/resource/ingressv2"
	"github.com/giantswarm/kvm-operator/service/resource/namespacev1"
	"github.com/giantswarm/kvm-operator/service/resource/namespacev2"
	"github.com/giantswarm/kvm-operator/service/resource/podv2"
	"github.com/giantswarm/kvm-operator/service/resource/pvcv1"
	"github.com/giantswarm/kvm-operator/service/resource/pvcv2"
	"github.com/giantswarm/kvm-operator/service/resource/servicev1"
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

	var ccService *cloudconfigv2.CloudConfig
	{
		c := cloudconfigv2.DefaultConfig()

		c.Logger = config.Logger

		ccService, err = cloudconfigv2.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configMapResource framework.Resource
	{
		c := configmapv2.DefaultConfig()

		c.CertWatcher = certWatcher
		c.CloudConfig = ccService
		c.K8sClient = k8sClient
		c.Logger = config.Logger

		configMapResource, err = configmapv2.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var deploymentResource framework.Resource
	{
		c := deploymentv2.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		deploymentResource, err = deploymentv2.New(c)
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

	// TODO remove after migration.
	migrateTPRsToCRDs(config.Logger, clientSet)

	var newWatcherFactory informer.WatcherFactory
	{
		newWatcherFactory = func() (watch.Interface, error) {
			watcher, err := clientSet.ProviderV1alpha1().KVMConfigs("").Watch(apismetav1.ListOptions{})
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

	var crdFramework *framework.Framework
	{
		c := framework.DefaultConfig()

		c.CRD = v1alpha1.NewKVMConfigCRD()
		c.CRDClient = crdClient
		c.Informer = newInformer
		c.InitCtxFunc = initCtxFunc
		c.Logger = config.Logger
		c.ResourceRouter = framework.DefaultResourceRouter(resources)

		crdFramework, err = framework.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return crdFramework, nil
}

func newCustomObjectFramework(config Config) (*framework.Framework, error) {
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

func migrateTPRsToCRDs(logger micrologger.Logger, clientSet *versioned.Clientset) {
	logger.Log("debug", "start TPR migration")

	var err error

	// List all TPOs.
	var b []byte
	{
		e := "/apis/cluster.giantswarm.io/v1/namespaces/default/kvms"
		b, err = clientSet.Discovery().RESTClient().Get().AbsPath(e).DoRaw()
		if err != nil {
			logger.Log("error", fmt.Sprintf("%#v", err))
			return
		}

		fmt.Printf("\n")
		fmt.Printf("b start\n")
		fmt.Printf("%s\n", b)
		fmt.Printf("b end\n")
		fmt.Printf("\n")
	}

	// Convert bytes into structure.
	var v *kvmtpr.List
	{
		if err := json.Unmarshal(b, v); err != nil {
			logger.Log("error", fmt.Sprintf("%#v", err))
			return
		}

		fmt.Printf("\n")
		fmt.Printf("v start\n")
		fmt.Printf("%#v\n", v)
		fmt.Printf("v end\n")
		fmt.Printf("\n")
	}

	// Iterate over all TPOs.
	for _, tpo := range v.Items {
		// Compute CRO using TPO.
		var cro *v1alpha1.KVMConfig
		{
			cro = &v1alpha1.KVMConfig{}

			cro.Spec.Cluster.Calico.CIDR = tpo.Spec.Cluster.Calico.CIDR
			cro.Spec.Cluster.Calico.Domain = tpo.Spec.Cluster.Calico.Domain
			cro.Spec.Cluster.Calico.MTU = tpo.Spec.Cluster.Calico.MTU
			cro.Spec.Cluster.Calico.Subnet = tpo.Spec.Cluster.Calico.Subnet
			cro.Spec.Cluster.Customer.ID = tpo.Spec.Cluster.Customer.ID
			cro.Spec.Cluster.Docker.Daemon.CIDR = tpo.Spec.Cluster.Docker.Daemon.CIDR
			cro.Spec.Cluster.Docker.Daemon.ExtraArgs = tpo.Spec.Cluster.Docker.Daemon.ExtraArgs
			cro.Spec.Cluster.Etcd.AltNames = tpo.Spec.Cluster.Etcd.AltNames
			cro.Spec.Cluster.Etcd.Domain = tpo.Spec.Cluster.Etcd.Domain
			cro.Spec.Cluster.Etcd.Port = tpo.Spec.Cluster.Etcd.Port
			cro.Spec.Cluster.Etcd.Prefix = tpo.Spec.Cluster.Etcd.Prefix
			cro.Spec.Cluster.ID = tpo.Spec.Cluster.Cluster.ID
			cro.Spec.Cluster.Kubernetes.API.AltNames = tpo.Spec.Cluster.Kubernetes.API.AltNames
			cro.Spec.Cluster.Kubernetes.API.ClusterIPRange = tpo.Spec.Cluster.Kubernetes.API.ClusterIPRange
			cro.Spec.Cluster.Kubernetes.API.Domain = tpo.Spec.Cluster.Kubernetes.API.Domain
			cro.Spec.Cluster.Kubernetes.API.InsecurePort = tpo.Spec.Cluster.Kubernetes.API.InsecurePort
			cro.Spec.Cluster.Kubernetes.API.IP = tpo.Spec.Cluster.Kubernetes.API.IP
			cro.Spec.Cluster.Kubernetes.API.SecurePort = tpo.Spec.Cluster.Kubernetes.API.SecurePort
			cro.Spec.Cluster.Kubernetes.DNS.IP = tpo.Spec.Cluster.Kubernetes.DNS.IP
			cro.Spec.Cluster.Kubernetes.Domain = tpo.Spec.Cluster.Kubernetes.Domain
			cro.Spec.Cluster.Kubernetes.Hyperkube.Docker.Image = tpo.Spec.Cluster.Kubernetes.Hyperkube.Docker.Image
			cro.Spec.Cluster.Kubernetes.IngressController.Docker.Image = tpo.Spec.Cluster.Kubernetes.IngressController.Docker.Image
			cro.Spec.Cluster.Kubernetes.IngressController.Domain = tpo.Spec.Cluster.Kubernetes.IngressController.Domain
			cro.Spec.Cluster.Kubernetes.IngressController.InsecurePort = tpo.Spec.Cluster.Kubernetes.IngressController.InsecurePort
			cro.Spec.Cluster.Kubernetes.IngressController.SecurePort = tpo.Spec.Cluster.Kubernetes.IngressController.SecurePort
			cro.Spec.Cluster.Kubernetes.IngressController.WildcardDomain = tpo.Spec.Cluster.Kubernetes.IngressController.WildcardDomain
			cro.Spec.Cluster.Kubernetes.Kubelet.AltNames = tpo.Spec.Cluster.Kubernetes.Kubelet.AltNames
			cro.Spec.Cluster.Kubernetes.Kubelet.Domain = tpo.Spec.Cluster.Kubernetes.Kubelet.Domain
			cro.Spec.Cluster.Kubernetes.Kubelet.Labels = tpo.Spec.Cluster.Kubernetes.Kubelet.Labels
			cro.Spec.Cluster.Kubernetes.Kubelet.Port = tpo.Spec.Cluster.Kubernetes.Kubelet.Port
			cro.Spec.Cluster.Kubernetes.NetworkSetup.Docker.Image = tpo.Spec.Cluster.Kubernetes.NetworkSetup.Docker.Image
			cro.Spec.Cluster.Kubernetes.SSH.UserList = toUserList(tpo.Spec.Cluster.Kubernetes.SSH.UserList)
			cro.Spec.Cluster.Masters = toClusterMasters(tpo.Spec.Cluster.Masters)
			cro.Spec.Cluster.Vault.Address = tpo.Spec.Cluster.Vault.Address
			cro.Spec.Cluster.Vault.Token = tpo.Spec.Cluster.Vault.Token
			cro.Spec.Cluster.Workers = toClusterWorkers(tpo.Spec.Cluster.Workers)
			cro.Spec.KVM.EndpointUpdater.Docker.Image = tpo.Spec.KVM.EndpointUpdater.Docker.Image
			cro.Spec.KVM.K8sKVM.Docker.Image = tpo.Spec.KVM.K8sKVM.Docker.Image
			cro.Spec.KVM.K8sKVM.StorageType = tpo.Spec.KVM.K8sKVM.StorageType
			cro.Spec.KVM.Masters = toKVMMasters(tpo.Spec.KVM.Masters)
			cro.Spec.KVM.Network.Flannel.VNI = tpo.Spec.KVM.Network.Flannel.VNI
			cro.Spec.KVM.NodeController.Docker.Image = tpo.Spec.KVM.NodeController.Docker.Image
			cro.Spec.KVM.Workers = toKVMWorkers(tpo.Spec.KVM.Workers)
			cro.Spec.VersionBundle.Version = tpo.Spec.VersionBundle.Version

			fmt.Printf("\n")
			fmt.Printf("cro start\n")
			fmt.Printf("%#v\n", cro)
			fmt.Printf("cro end\n")
			fmt.Printf("\n")
		}

		// TODO create CRO in Kubernetes API.
	}

	logger.Log("debug", "end TPR migration")
}

func toClusterMasters(masters []spec.Node) []v1alpha1.ClusterNode {
	var newList []v1alpha1.ClusterNode

	for _, master := range masters {
		n := v1alpha1.ClusterNode{
			ID: master.ID,
		}

		newList = append(newList, n)
	}

	return newList
}

func toClusterWorkers(workers []spec.Node) []v1alpha1.ClusterNode {
	var newList []v1alpha1.ClusterNode

	for _, worker := range workers {
		n := v1alpha1.ClusterNode{
			ID: worker.ID,
		}

		newList = append(newList, n)
	}

	return newList
}

func toKVMMasters(masters []kvm.Node) []v1alpha1.KVMConfigSpecKVMNode {
	var newList []v1alpha1.KVMConfigSpecKVMNode

	for _, master := range masters {
		w := v1alpha1.KVMConfigSpecKVMNode{
			CPUs:   master.CPUs,
			Disk:   master.Disk,
			Memory: master.Memory,
		}

		newList = append(newList, w)
	}

	return newList
}

func toKVMWorkers(workers []kvm.Node) []v1alpha1.KVMConfigSpecKVMNode {
	var newList []v1alpha1.KVMConfigSpecKVMNode

	for _, worker := range workers {
		w := v1alpha1.KVMConfigSpecKVMNode{
			CPUs:   worker.CPUs,
			Disk:   worker.Disk,
			Memory: worker.Memory,
		}

		newList = append(newList, w)
	}

	return newList
}

func toUserList(userList []ssh.User) []v1alpha1.ClusterKubernetesSSHUser {
	var newList []v1alpha1.ClusterKubernetesSSHUser

	for _, user := range userList {
		u := v1alpha1.ClusterKubernetesSSHUser{
			Name:      user.Name,
			PublicKey: user.PublicKey,
		}

		newList = append(newList, u)
	}

	return newList
}
