package kvmconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/context/updateallowedcontext"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/randomkeys"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/service/kvmconfig/v2"
	v2key "github.com/giantswarm/kvm-operator/service/kvmconfig/v2/key"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v3"
	v3key "github.com/giantswarm/kvm-operator/service/kvmconfig/v3/key"
)

type FrameworkConfig struct {
	G8sClient    versioned.Interface
	K8sClient    kubernetes.Interface
	K8sExtClient apiextensionsclient.Interface
	Logger       micrologger.Logger

	GuestUpdateEnabled bool
	// Name is the name of the project.
	Name string
}

func NewFramework(config FrameworkConfig) (*framework.Framework, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.G8sClient must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.K8sExtClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sExtClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.Name == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Name must not be empty")
	}

	var err error

	var crdClient *k8scrdclient.CRDClient
	{
		c := k8scrdclient.DefaultConfig()

		c.K8sExtClient = config.K8sExtClient
		c.Logger = config.Logger

		crdClient, err = k8scrdclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var certsSearcher certs.Interface
	{
		c := certs.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		certsSearcher, err = certs.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var randomkeysSearcher randomkeys.Interface
	{
		keyConfig := randomkeys.DefaultConfig()
		keyConfig.K8sClient = config.K8sClient
		keyConfig.Logger = config.Logger
		randomkeysSearcher, err = randomkeys.NewSearcher(keyConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	v2ResourceSetHandles := func(obj interface{}) bool {
		customObject, err := v2key.ToCustomObject(obj)
		if err != nil {
			return false
		}
		versionBundleVersion := v2key.VersionBundleVersion(customObject)

		if versionBundleVersion == "1.0.0" {
			return true
		}
		if versionBundleVersion == "0.1.0" {
			return true
		}
		if versionBundleVersion == "" {
			return true
		}

		return false
	}

	var v2Resources []framework.Resource
	{
		c := v2.ResourcesConfig{
			CertsSearcher:      certsSearcher,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			Name: config.Name,
		}

		v2Resources, err = v2.NewResources(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	v3ResourceSetHandles := func(obj interface{}) bool {
		customObject, err := v3key.ToCustomObject(obj)
		if err != nil {
			return false
		}
		versionBundleVersion := v3key.VersionBundleVersion(customObject)

		if versionBundleVersion == "1.1.0" {
			return true
		}

		return false
	}

	v3InitCtx := func(ctx context.Context, obj interface{}) (context.Context, error) {
		if config.GuestUpdateEnabled {
			updateallowedcontext.SetUpdateAllowed(ctx)
		}

		return ctx, nil
	}

	var v3Resources []framework.Resource
	{
		c := v3.ResourcesConfig{
			CertsSearcher:      certsSearcher,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomkeysSearcher,

			Name: config.Name,
		}

		v3Resources, err = v3.NewResources(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var newInformer *informer.Informer
	{
		c := informer.DefaultConfig()

		c.Watcher = config.G8sClient.ProviderV1alpha1().KVMConfigs("")

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v2ResourceSet *framework.ResourceSet
	{
		c := framework.ResourceSetConfig{}

		c.Handles = v2ResourceSetHandles
		c.Logger = config.Logger
		c.Resources = v2Resources

		v2ResourceSet, err = framework.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v3ResourceSet *framework.ResourceSet
	{
		c := framework.ResourceSetConfig{}

		c.Handles = v3ResourceSetHandles
		c.InitCtx = v3InitCtx
		c.Logger = config.Logger
		c.Resources = v3Resources

		v3ResourceSet, err = framework.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceRouter *framework.ResourceRouter
	{
		c := framework.ResourceRouterConfig{}

		c.ResourceSets = []*framework.ResourceSet{
			v2ResourceSet,
			v3ResourceSet,
		}

		resourceRouter, err = framework.NewResourceRouter(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var crdFramework *framework.Framework
	{
		c := framework.DefaultConfig()

		c.CRD = v1alpha1.NewKVMConfigCRD()
		c.CRDClient = crdClient
		c.Informer = newInformer
		c.Logger = config.Logger
		c.ResourceRouter = resourceRouter

		crdFramework, err = framework.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return crdFramework, nil
}
