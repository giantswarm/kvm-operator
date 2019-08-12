package controller

import (
	"time"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/tenantcluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/service/controller/v20"
	"github.com/giantswarm/kvm-operator/service/controller/v21"
	"github.com/giantswarm/kvm-operator/service/controller/v22"
	"github.com/giantswarm/kvm-operator/service/controller/v23"
	"github.com/giantswarm/kvm-operator/service/controller/v23patch1"
	"github.com/giantswarm/kvm-operator/service/controller/v24"
)

type DeleterConfig struct {
	CertsSearcher certs.Interface
	G8sClient     versioned.Interface
	K8sClient     kubernetes.Interface
	Logger        micrologger.Logger
	TenantCluster tenantcluster.Interface

	CRDLabelSelector string
	ProjectName      string
}

func (c DeleterConfig) newInformerListOptions() metav1.ListOptions {
	listOptions := metav1.ListOptions{
		LabelSelector: c.CRDLabelSelector,
	}

	return listOptions
}

type Deleter struct {
	*controller.Controller
}

func NewDeleter(config DeleterConfig) (*Deleter, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}

	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}

	var err error

	var newInformer *informer.Informer
	{
		c := informer.Config{
			Logger:  config.Logger,
			Watcher: config.G8sClient.ProviderV1alpha1().KVMConfigs(""),

			ListOptions:  config.newInformerListOptions(),
			RateWait:     informer.DefaultRateWait,
			ResyncPeriod: 30 * time.Second,
		}

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets, err := newDeleterResourceSets(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			Informer:     newInformer,
			Logger:       config.Logger,
			ResourceSets: resourceSets,
			RESTClient:   config.G8sClient.ProviderV1alpha1().RESTClient(),

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

func newDeleterResourceSets(config DeleterConfig) ([]*controller.ResourceSet, error) {
	var err error

	var resourceSetV20 *controller.ResourceSet
	{
		c := v20.DeleterResourceSetConfig{
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,
			TenantCluster: config.TenantCluster,

			ProjectName: config.ProjectName,
		}

		resourceSetV20, err = v20.NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV21 *controller.ResourceSet
	{
		c := v21.DeleterResourceSetConfig{
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,
			TenantCluster: config.TenantCluster,

			ProjectName: config.ProjectName,
		}

		resourceSetV21, err = v21.NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV22 *controller.ResourceSet
	{
		c := v22.DeleterResourceSetConfig{
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,
			TenantCluster: config.TenantCluster,

			ProjectName: config.ProjectName,
		}

		resourceSetV22, err = v22.NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV23 *controller.ResourceSet
	{
		c := v23.DeleterResourceSetConfig{
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,
			TenantCluster: config.TenantCluster,

			ProjectName: config.ProjectName,
		}

		resourceSetV23, err = v23.NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV23patch1 *controller.ResourceSet
	{
		c := v23patch1.DeleterResourceSetConfig{
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,
			TenantCluster: config.TenantCluster,

			ProjectName: config.ProjectName,
		}

		resourceSetV23patch1, err = v23patch1.NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV24 *controller.ResourceSet
	{
		c := v24.DeleterResourceSetConfig{
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,
			TenantCluster: config.TenantCluster,

			ProjectName: config.ProjectName,
		}

		resourceSetV24, err = v24.NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		resourceSetV20,
		resourceSetV21,
		resourceSetV22,
		resourceSetV23,
		resourceSetV23patch1,
		resourceSetV24,
	}

	return resourceSets, nil
}
