package controller

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/tenantcluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type DeleterConfig struct {
	CertsSearcher certs.Interface
	K8sClient     k8sclient.Interface
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
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	var err error

	resourceSets, err := newDeleterResourceSets(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			CRD:          v1alpha1.NewKVMConfigCRD(),
			K8sClient:    config.K8sClient,
			Logger:       config.Logger,
			ResourceSets: resourceSets,
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

func newDeleterResourceSets(config DeleterConfig) ([]*controller.ResourceSet, error) {
	var err error

	var resourceSet *controller.ResourceSet
	{
		c := DeleterResourceSetConfig{
			K8sClient:     config.K8sClient.K8sClient(),
			Logger:        config.Logger,
			TenantCluster: config.TenantCluster,

			ProjectName: config.ProjectName,
		}

		resourceSet, err = NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		resourceSet,
	}

	return resourceSets, nil
}
