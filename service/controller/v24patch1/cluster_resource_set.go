<<<<<<< HEAD
<<<<<<< HEAD
package v24patch1
=======
package v24
>>>>>>> c4c6c79d... copy v24 to v24patch1
=======
package v24patch1
>>>>>>> d6f149c2... wire v24patch1

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/context/updateallowedcontext"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/crud"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	"github.com/giantswarm/randomkeys"
	"github.com/giantswarm/statusresource"
	"github.com/giantswarm/tenantcluster"
	"k8s.io/client-go/kubernetes"

<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> d6f149c2... wire v24patch1
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/key"
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/resource/clusterrolebinding"
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/resource/configmap"
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/resource/deployment"
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/resource/ingress"
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/resource/namespace"
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/resource/nodeindexstatus"
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/resource/pvc"
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/resource/service"
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/resource/serviceaccount"
<<<<<<< HEAD
=======
	"github.com/giantswarm/kvm-operator/service/controller/v24/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/controller/v24/key"
	"github.com/giantswarm/kvm-operator/service/controller/v24/resource/clusterrolebinding"
	"github.com/giantswarm/kvm-operator/service/controller/v24/resource/configmap"
	"github.com/giantswarm/kvm-operator/service/controller/v24/resource/deployment"
	"github.com/giantswarm/kvm-operator/service/controller/v24/resource/ingress"
	"github.com/giantswarm/kvm-operator/service/controller/v24/resource/namespace"
	"github.com/giantswarm/kvm-operator/service/controller/v24/resource/nodeindexstatus"
	"github.com/giantswarm/kvm-operator/service/controller/v24/resource/pvc"
	"github.com/giantswarm/kvm-operator/service/controller/v24/resource/service"
	"github.com/giantswarm/kvm-operator/service/controller/v24/resource/serviceaccount"
>>>>>>> c4c6c79d... copy v24 to v24patch1
=======
>>>>>>> d6f149c2... wire v24patch1
)

type ClusterResourceSetConfig struct {
	CertsSearcher      certs.Interface
	G8sClient          versioned.Interface
	K8sClient          kubernetes.Interface
	Logger             micrologger.Logger
	RandomkeysSearcher randomkeys.Interface
	TenantCluster      tenantcluster.Interface

	DNSServers         string
	GuestUpdateEnabled bool
	IgnitionPath       string
	NTPServers         string
	OIDC               cloudconfig.OIDCConfig
	ProjectName        string
	SSOPublicKey       string
}

func NewClusterResourceSet(config ClusterResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var cloudConfig *cloudconfig.CloudConfig
	{
		c := cloudconfig.Config{
			Logger: config.Logger,

			IgnitionPath: config.IgnitionPath,
			OIDC:         config.OIDC,
			SSOPublicKey: config.SSOPublicKey,
		}

		cloudConfig, err = cloudconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterRoleBindingResource resource.Interface
	{
		c := clusterrolebinding.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		ops, err := clusterrolebinding.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		clusterRoleBindingResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var namespaceResource resource.Interface
	{
		c := namespace.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		ops, err := namespace.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		namespaceResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var serviceAccountResource resource.Interface
	{
		c := serviceaccount.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		ops, err := serviceaccount.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		serviceAccountResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configMapResource resource.Interface
	{
		c := configmap.Config{
			CertsSearcher: config.CertsSearcher,
			CloudConfig:   cloudConfig,
			K8sClient:     config.K8sClient,
			KeyWatcher:    config.RandomkeysSearcher,
			Logger:        config.Logger,
		}

		ops, err := configmap.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMapResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var deploymentResource resource.Interface
	{
		c := deployment.Config{
			DNSServers: config.DNSServers,
			K8sClient:  config.K8sClient,
			Logger:     config.Logger,
			NTPServers: config.NTPServers,
		}

		ops, err := deployment.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		deploymentResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ingressResource resource.Interface
	{
		c := ingress.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		ops, err := ingress.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		ingressResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var nodeIndexStatusResource resource.Interface
	{
		c := nodeindexstatus.Config{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
		}

		nodeIndexStatusResource, err = nodeindexstatus.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var pvcResource resource.Interface
	{
		c := pvc.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		ops, err := pvc.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		pvcResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var serviceResource resource.Interface
	{
		c := service.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		ops, err := service.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		serviceResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var statusResource resource.Interface
	{
		c := statusresource.ResourceConfig{
			ClusterEndpointFunc:      key.ToClusterEndpoint,
			ClusterIDFunc:            key.ToClusterID,
			ClusterStatusFunc:        key.ToClusterStatus,
			NodeCountFunc:            key.ToNodeCount,
			Logger:                   config.Logger,
			RESTClient:               config.G8sClient.ProviderV1alpha1().RESTClient(),
			TenantCluster:            config.TenantCluster,
			VersionBundleVersionFunc: key.ToVersionBundleVersion,
		}

		statusResource, err = statusresource.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		statusResource,
		nodeIndexStatusResource,
		clusterRoleBindingResource,
		namespaceResource,
		serviceAccountResource,
		configMapResource,
		deploymentResource,
		ingressResource,
		pvcResource,
		serviceResource,
	}

	{
		c := retryresource.WrapConfig{
			Logger: config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{}

		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	handlesFunc := func(obj interface{}) bool {
		kvmConfig, err := key.ToCustomObject(obj)
		if err != nil {
			return false
		}

		if key.VersionBundleVersion(kvmConfig) == VersionBundle().Version {
			return true
		}

		return false
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		if config.GuestUpdateEnabled {
			updateallowedcontext.SetUpdateAllowed(ctx)
		}

		return ctx, nil
	}

	var clusterResourceSet *controller.ResourceSet
	{
		c := controller.ResourceSetConfig{
			Handles:   handlesFunc,
			InitCtx:   initCtxFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		clusterResourceSet, err = controller.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return clusterResourceSet, nil
}

func toCRUDResource(logger micrologger.Logger, ops crud.Interface) (resource.Interface, error) {
	c := crud.ResourceConfig{
		CRUD:   ops,
		Logger: logger,
	}

	r, err := crud.NewResource(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return r, nil
}
