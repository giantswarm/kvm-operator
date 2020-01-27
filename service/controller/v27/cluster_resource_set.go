package v27

import (
	"context"

	"github.com/giantswarm/certs"
	"github.com/giantswarm/k8sclient"
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

	"github.com/giantswarm/kvm-operator/service/controller/v27/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/controller/v27/key"
	"github.com/giantswarm/kvm-operator/service/controller/v27/resource/clusterrolebinding"
	"github.com/giantswarm/kvm-operator/service/controller/v27/resource/configmap"
	"github.com/giantswarm/kvm-operator/service/controller/v27/resource/deployment"
	"github.com/giantswarm/kvm-operator/service/controller/v27/resource/ingress"
	"github.com/giantswarm/kvm-operator/service/controller/v27/resource/namespace"
	"github.com/giantswarm/kvm-operator/service/controller/v27/resource/nodeindexstatus"
	"github.com/giantswarm/kvm-operator/service/controller/v27/resource/pvc"
	"github.com/giantswarm/kvm-operator/service/controller/v27/resource/service"
	"github.com/giantswarm/kvm-operator/service/controller/v27/resource/serviceaccount"
)

type ClusterResourceSetConfig struct {
	CertsSearcher      certs.Interface
	K8sClient          k8sclient.Interface
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
			K8sClient: config.K8sClient.K8sClient(),
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
		c := namespace.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

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
		c := serviceaccount.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

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
			K8sClient:     config.K8sClient.K8sClient(),
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
			K8sClient:  config.K8sClient.K8sClient(),
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
		c := ingress.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

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
			G8sClient: config.K8sClient.G8sClient(),
			Logger:    config.Logger,
		}

		nodeIndexStatusResource, err = nodeindexstatus.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var pvcResource resource.Interface
	{
		c := pvc.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

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
		c := service.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

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
			RESTClient:               config.K8sClient.G8sClient().ProviderV1alpha1().RESTClient(),
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
