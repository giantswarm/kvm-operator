// +build k8srequired

package teardown

import (
	"context"
	"fmt"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/integration/utils"
)

// Teardown e2e testing environment.
func Teardown(g *framework.Guest, h *framework.Host) error {
	var err error
	var l micrologger.Logger
	{
		c := micrologger.Config{}

		l, err = micrologger.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}
	clusterID := getClusterID(h.TargetNamespace())
	ctx := context.Background()

	// get flannel info so we can delete it from rangepool
	var flannelNetwork string
	{
		flannelConfig, err := h.G8sClient().CoreV1alpha1().FlannelConfigs(v1.NamespaceDefault).Get(clusterID, v1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
		flannelNetwork = flannelConfig.Spec.Flannel.Spec.Network
	}

	{
		err = framework.HelmCmd(fmt.Sprintf("delete %s-cert-config-e2e --purge", h.TargetNamespace()))
		if err != nil {
			return microerror.Mask(err)
		}
		err = framework.HelmCmd(fmt.Sprintf("delete %s-flannel-config-e2e --purge", h.TargetNamespace()))
		if err != nil {
			return microerror.Mask(err)
		}
		err = framework.HelmCmd(fmt.Sprintf("delete %s-kvm-config-e2e --purge", h.TargetNamespace()))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// wait until crds are deleted by operators
	o := func() error {
		certList, err := h.G8sClient().CoreV1alpha1().CertConfigs(v1.NamespaceDefault).List(v1.ListOptions{
			LabelSelector: crdLabelSelector(clusterID),
		})
		if err != nil {
			return microerror.Mask(err)
		}
		if len(certList.Items) == 0 {
			// resource doesnt exist, we are good to continue
			l.Log("level", "info", "message", "cert crd was deleted")
		} else {
			l.Log("level", "info", "message", "cert crd has not been deleted")
			return microerror.Mask(resourceNotDeleted)
		}

		flannelList, err := h.G8sClient().CoreV1alpha1().FlannelConfigs(v1.NamespaceDefault).List(v1.ListOptions{
			LabelSelector: crdLabelSelector(clusterID),
		})

		if err != nil {
			return microerror.Mask(err)
		}
		if len(flannelList.Items) == 0 {
			// resource doesnt exist, we are good to continue
			l.Log("level", "info", "message", "flannel crd was deleted")
		} else {
			l.Log("level", "info", "message", "flannel crd has not been deleted")
			return microerror.Mask(resourceNotDeleted)
		}

		kvmList, err := h.G8sClient().ProviderV1alpha1().KVMConfigs(v1.NamespaceDefault).List(v1.ListOptions{
			LabelSelector: crdLabelSelector(clusterID),
		})
		if err != nil {
			return microerror.Mask(err)
		}

		if len(kvmList.Items) == 0 {
			// resource doesnt exist, we are good to continue
			l.Log("level", "info", "message", "kvm crd was deleted")
		} else {
			l.Log("level", "info", "message", "kvm crd has not been deleted")
			return microerror.Mask(resourceNotDeleted)
		}
		return nil
	}
	b := backoff.NewExponential(framework.LongMaxWait, framework.LongMaxInterval)
	n := backoff.NewNotifier(l, context.Background())
	err = backoff.RetryNotify(o, b, n)

	{
		err = h.K8sClient().CoreV1().Namespaces().Delete(h.TargetNamespace(), &v1.DeleteOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// clear rangepool values
	{
		crdStorage, err := utils.InitCRDStorage(ctx, h, l)
		if err != nil {
			return microerror.Mask(err)
		}
		rangePool, err := utils.InitRangePool(crdStorage, l)
		if err != nil {
			return microerror.Mask(err)
		}

		err = utils.DeleteVNI(ctx, rangePool, clusterID)
		if err != nil {
			return microerror.Mask(err)
		}
		l.Log("level", "info", "message", "Deleted VNI reservation in rangepool.")
		err = utils.DeleteIngressNodePorts(ctx, rangePool, clusterID)
		if err != nil {
			return microerror.Mask(err)
		}
		l.Log("level", "info", "message", "Deleted Ingress node port reservation in rangepool.")
		err = utils.DeleteFlannelNetwork(ctx, flannelNetwork, crdStorage, l)
		if err != nil {
			return microerror.Mask(err)
		}

	}

	return nil
}
