package controller

import (
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/tenantcluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	v20 "github.com/giantswarm/kvm-operator/service/controller/v20"
	v21 "github.com/giantswarm/kvm-operator/service/controller/v21"
	v22 "github.com/giantswarm/kvm-operator/service/controller/v22"
	v23 "github.com/giantswarm/kvm-operator/service/controller/v23"
	"github.com/giantswarm/kvm-operator/service/controller/v23patch1"
	v24 "github.com/giantswarm/kvm-operator/service/controller/v24"
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1"
	v25 "github.com/giantswarm/kvm-operator/service/controller/v25"
	v26 "github.com/giantswarm/kvm-operator/service/controller/v26"
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
			ResyncPeriod: deleterResyncPeriod,
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

	var resourceSetV20 *controller.ResourceSet
	{
		c := v20.DeleterResourceSetConfig{
			K8sClient:     config.K8sClient.K8sClient(),
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
			K8sClient:     config.K8sClient.K8sClient(),
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
			K8sClient:     config.K8sClient.K8sClient(),
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
			K8sClient:     config.K8sClient.K8sClient(),
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
			K8sClient:     config.K8sClient.K8sClient(),
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
			K8sClient:     config.K8sClient.K8sClient(),
			Logger:        config.Logger,
			TenantCluster: config.TenantCluster,

			ProjectName: config.ProjectName,
		}

		resourceSetV24, err = v24.NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV24patch1 *controller.ResourceSet
	{
		c := v24patch1.DeleterResourceSetConfig{
			K8sClient:     config.K8sClient.K8sClient(),
			Logger:        config.Logger,
			TenantCluster: config.TenantCluster,

			ProjectName: config.ProjectName,
		}

		resourceSetV24patch1, err = v24patch1.NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV25 *controller.ResourceSet
	{
		c := v25.DeleterResourceSetConfig{
			K8sClient:     config.K8sClient.K8sClient(),
			Logger:        config.Logger,
			TenantCluster: config.TenantCluster,

			ProjectName: config.ProjectName,
		}

		resourceSetV25, err = v25.NewDeleterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV26 *controller.ResourceSet
	{
		c := v26.DeleterResourceSetConfig{
			K8sClient:     config.K8sClient.K8sClient(),
			Logger:        config.Logger,
			TenantCluster: config.TenantCluster,

			ProjectName: config.ProjectName,
		}

		resourceSetV26, err = v26.NewDeleterResourceSet(c)
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
		resourceSetV24patch1,
		resourceSetV25,
		resourceSetV26,
	}

	return resourceSets, nil
}
