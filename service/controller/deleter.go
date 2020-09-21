package controller

import (
	"time"

	"github.com/giantswarm/apiextensions/v2/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/v3/pkg/certs"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v2/pkg/controller"
	"github.com/giantswarm/operatorkit/v2/pkg/resource"
	"github.com/giantswarm/operatorkit/v2/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v2/pkg/resource/wrapper/retryresource"
	"github.com/giantswarm/tenantcluster/v3/pkg/tenantcluster"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/kvm-operator/pkg/label"
	"github.com/giantswarm/kvm-operator/pkg/project"
	"github.com/giantswarm/kvm-operator/service/controller/resource/cleanupendpointips"
	"github.com/giantswarm/kvm-operator/service/controller/resource/node"
)

const deleterResyncPeriod = time.Minute * 2

type DeleterConfig struct {
	CertsSearcher certs.Interface
	K8sClient     k8sclient.Interface
	Logger        micrologger.Logger
	TenantCluster tenantcluster.Interface

	CRDLabelSelector string
	ProjectName      string
}

type Deleter struct {
	*controller.Controller
}

func NewDeleter(config DeleterConfig) (*Deleter, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	var err error

	resources, err := newDeleterResources(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			K8sClient:    config.K8sClient,
			Logger:       config.Logger,
			Resources:    resources,
			ResyncPeriod: deleterResyncPeriod,
			Selector: labels.SelectorFromSet(map[string]string{
				label.OperatorVersion: project.Version(),
			}),
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(v1alpha1.KVMConfig)
			},

			Name: config.ProjectName + "-deleter",
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	d := &Deleter{
		Controller: operatorkitController,
	}

	return d, nil
}

func newDeleterResources(config DeleterConfig) ([]resource.Interface, error) {
	var err error

	var cleanupendpointipsResource resource.Interface
	{
		c := cleanupendpointips.Config{
			K8sClient:     config.K8sClient.K8sClient(),
			Logger:        config.Logger,
			TenantCluster: config.TenantCluster,
		}

		cleanupendpointipsResource, err = cleanupendpointips.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	var nodeResource resource.Interface
	{
		c := node.Config{
			K8sClient:     config.K8sClient.K8sClient(),
			Logger:        config.Logger,
			TenantCluster: config.TenantCluster,
		}

		nodeResource, err = node.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		cleanupendpointipsResource,
		nodeResource,
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
