package controller

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v4/pkg/resource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/crud"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/retryresource"
	"github.com/giantswarm/randomkeys/v2"
	"github.com/giantswarm/statusresource/v3"

	"github.com/giantswarm/kvm-operator/service/controller/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/controller/key"
	"github.com/giantswarm/kvm-operator/service/controller/resource/clusterrolebinding"
	"github.com/giantswarm/kvm-operator/service/controller/resource/configmap"
	"github.com/giantswarm/kvm-operator/service/controller/resource/deployment"
	"github.com/giantswarm/kvm-operator/service/controller/resource/ingress"
	"github.com/giantswarm/kvm-operator/service/controller/resource/namespace"
	"github.com/giantswarm/kvm-operator/service/controller/resource/nodecontroller"
	"github.com/giantswarm/kvm-operator/service/controller/resource/nodeindexstatus"
	"github.com/giantswarm/kvm-operator/service/controller/resource/pvc"
	"github.com/giantswarm/kvm-operator/service/controller/resource/service"
	"github.com/giantswarm/kvm-operator/service/controller/resource/serviceaccount"
)

func newClusterResources(config ClusterConfig) ([]resource.Interface, error) {
	var err error

	var randomkeysSearcher randomkeys.Interface
	{
		c := randomkeys.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

		randomkeysSearcher, err = randomkeys.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cloudConfig *cloudconfig.CloudConfig
	{
		c := cloudconfig.Config{
			Logger: config.Logger,

			DockerhubToken: config.DockerhubToken,
			IgnitionPath:   config.IgnitionPath,
			OIDC: cloudconfig.OIDCConfig{
				ClientID:       config.OIDC.ClientID,
				IssuerURL:      config.OIDC.IssuerURL,
				UsernameClaim:  config.OIDC.UsernameClaim,
				UsernamePrefix: config.OIDC.UsernamePrefix,
				GroupsClaim:    config.OIDC.GroupsClaim,
				GroupsPrefix:   config.OIDC.GroupsPrefix,
			},
			RegistryMirrors: config.RegistryMirrors,
			SSOPublicKey:    config.SSOPublicKey,
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

			ClusterRoleGeneral: config.ClusterRoleGeneral,
			ClusterRolePSP:     config.ClusterRolePSP,
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
			CertsSearcher:  config.CertsSearcher,
			CloudConfig:    cloudConfig,
			G8sClient:      config.K8sClient.G8sClient(),
			K8sClient:      config.K8sClient.K8sClient(),
			KeyWatcher:     randomkeysSearcher,
			Logger:         config.Logger,
			RegistryDomain: config.RegistryDomain,
			DockerhubToken: config.DockerhubToken,
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
			DNSServers:    config.DNSServers,
			G8sClient:     config.K8sClient.G8sClient(),
			K8sClient:     config.K8sClient.K8sClient(),
			Logger:        config.Logger,
			NTPServers:    config.NTPServers,
			TenantCluster: config.TenantCluster,
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
			VersionBundleVersionFunc: key.ToOperatorVersion,
		}

		statusResource, err = statusresource.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var nodeControllerResource resource.Interface
	{
		c := nodecontroller.Config{
			K8sClient:       config.K8sClient,
			Logger:          config.Logger,
			WorkloadCluster: config.TenantCluster,
		}

		nodeControllerResource, err = nodecontroller.New(c)
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
		nodeControllerResource,
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

	return resources, nil
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
