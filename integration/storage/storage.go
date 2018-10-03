package storage

import (
	"context"

	"github.com/giantswarm/crdstorage"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	giantswarmNamespace = "giantswarm"
)

func InitCRDStorage(ctx context.Context, h *framework.Host, l micrologger.Logger) (microstorage.Storage, error) {
	var err error

	k8sExtClient, err := apiextensionsclient.NewForConfig(h.RestConfig())
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var k8sCrdClient *k8scrdclient.CRDClient
	{
		var k8sCrdClientConfig k8scrdclient.Config
		k8sCrdClientConfig.Logger = l
		k8sCrdClientConfig.K8sExtClient = k8sExtClient

		k8sCrdClient, err = k8scrdclient.New(k8sCrdClientConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := crdstorage.DefaultConfig()
	c.CRDClient = k8sCrdClient
	c.G8sClient = h.G8sClient()
	c.K8sClient = h.K8sClient()
	c.Logger = l

	targetNamespace := h.TargetNamespace()
	if targetNamespace == "" {
		targetNamespace = giantswarmNamespace
	}

	c.Name = "kvm-e2e"
	c.Namespace = &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: targetNamespace,
		},
	}

	crdStorage, err := crdstorage.New(c)

	if err != nil {
		return nil, microerror.Mask(err)
	}

	l.LogCtx(ctx, "info", "booting crdstorage")
	err = crdStorage.Boot(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return crdStorage, nil
}
