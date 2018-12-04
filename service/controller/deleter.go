package controller

import (
	"time"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/informer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/service/controller/v13"
	"github.com/giantswarm/kvm-operator/service/controller/v14"
	"github.com/giantswarm/kvm-operator/service/controller/v14patch1"
	"github.com/giantswarm/kvm-operator/service/controller/v14patch2"
	"github.com/giantswarm/kvm-operator/service/controller/v15"
	"github.com/giantswarm/kvm-operator/service/controller/v16"
	"github.com/giantswarm/kvm-operator/service/controller/v17"
)

type DeleterConfig struct {
	CertsSearcher certs.Interface
	G8sClient     versioned.Interface
	K8sClient     kubernetes.Interface
	Logger        micrologger.Logger

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

	var resourceSetV13 *controller.ResourceSet
	{
		c := v13.DeleterResourceSetConfig{
			CertsSearcher: config.CertsSearcher,
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,

			ProjectName: config.ProjectName,
		}

		resourceSetV13, err = v13.NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV14 *controller.ResourceSet
	{
		c := v14.DeleterResourceSetConfig{
			CertsSearcher: config.CertsSearcher,
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,

			ProjectName: config.ProjectName,
		}

		resourceSetV14, err = v14.NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV14Patch1 *controller.ResourceSet
	{
		c := v14patch1.DeleterResourceSetConfig{
			CertsSearcher: config.CertsSearcher,
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,

			ProjectName: config.ProjectName,
		}

		resourceSetV14Patch1, err = v14patch1.NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV14Patch2 *controller.ResourceSet
	{
		c := v14patch2.DeleterResourceSetConfig{
			CertsSearcher: config.CertsSearcher,
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,

			ProjectName: config.ProjectName,
		}

		resourceSetV14Patch2, err = v14patch2.NewDeleterResourceSet(c)
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

	resourceSets := []*controller.ResourceSet{
		resourceSetV13,
		resourceSetV14,
		resourceSetV14Patch1,
		resourceSetV14Patch2,
		resourceSetV15,
		resourceSetV16,
		resourceSetV17,
	}

	return resourceSets, nil
}
