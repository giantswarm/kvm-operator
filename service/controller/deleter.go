package controller

import (
	"github.com/giantswarm/kvm-operator/service/controller/v19"
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

	"github.com/giantswarm/kvm-operator/service/controller/v14patch3"
	"github.com/giantswarm/kvm-operator/service/controller/v14patch4"
	"github.com/giantswarm/kvm-operator/service/controller/v15"
	"github.com/giantswarm/kvm-operator/service/controller/v16"
	"github.com/giantswarm/kvm-operator/service/controller/v17"
	"github.com/giantswarm/kvm-operator/service/controller/v17patch1"
	"github.com/giantswarm/kvm-operator/service/controller/v18"
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

	var resourceSetV14Patch3 *controller.ResourceSet
	{
		c := v14patch3.DeleterResourceSetConfig{
			CertsSearcher: config.CertsSearcher,
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,

			ProjectName: config.ProjectName,
		}

		resourceSetV14Patch3, err = v14patch3.NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV14Patch4 *controller.ResourceSet
	{
		c := v14patch4.DeleterResourceSetConfig{
			CertsSearcher: config.CertsSearcher,
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,

			ProjectName: config.ProjectName,
		}

		resourceSetV14Patch4, err = v14patch4.NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV15 *controller.ResourceSet
	{
		c := v15.DeleterResourceSetConfig{
			CertsSearcher: config.CertsSearcher,
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,

			ProjectName: config.ProjectName,
		}

		resourceSetV15, err = v15.NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV16 *controller.ResourceSet
	{
		c := v16.DeleterResourceSetConfig{
			CertsSearcher: config.CertsSearcher,
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,

			ProjectName: config.ProjectName,
		}

		resourceSetV16, err = v16.NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV17 *controller.ResourceSet
	{
		c := v17.DeleterResourceSetConfig{
			CertsSearcher: config.CertsSearcher,
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,

			ProjectName: config.ProjectName,
		}

		resourceSetV17, err = v17.NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV17patch1 *controller.ResourceSet
	{
		c := v17patch1.DeleterResourceSetConfig{
			CertsSearcher: config.CertsSearcher,
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,

			ProjectName: config.ProjectName,
		}

		resourceSetV17patch1, err = v17patch1.NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV18 *controller.ResourceSet
	{
		c := v18.DeleterResourceSetConfig{
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,
			TenantCluster: config.TenantCluster,

			ProjectName: config.ProjectName,
		}

		resourceSetV18, err = v18.NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV19 *controller.ResourceSet
	{
		c := v19.DeleterResourceSetConfig{
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,
			TenantCluster: config.TenantCluster,

			ProjectName: config.ProjectName,
		}

		resourceSetV19, err = v19.NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		resourceSetV14Patch3,
		resourceSetV14Patch4,
		resourceSetV15,
		resourceSetV16,
		resourceSetV17,
		resourceSetV17patch1,
		resourceSetV18,
		resourceSetV19,
	}

	return resourceSets, nil
}
