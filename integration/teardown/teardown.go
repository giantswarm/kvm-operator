// +build k8srequired

package teardown

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
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

	{
		err = framework.HelmCmd(fmt.Sprintf("delete %s-cert-config-e2e --purge", h.TargetNamespace()))
		if err != nil {
			return microerror.Mask(err)
		}

		err = framework.HelmCmd(fmt.Sprintf("delete %s-kvm-config-e2e --purge", h.TargetNamespace()))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// wait until both crds are deleted by operators
	o := func() error {
		kvmDeleted, certDeleted := false, false
		kvmList, err := h.G8sClient().ProviderV1alpha1().KVMConfigs(v1.NamespaceDefault).List(v1.ListOptions{
			LabelSelector: crdLabelSelector(h.TargetNamespace()),
		})
		if err != nil {
			return microerror.Mask(err)
		}

		if len(kvmList.Items) == 0 {
			// resource doesnt exist we are good to continue
			kvmDeleted = true
			l.Log("level", "info", "message", "kvm crd was deleted")
		} else {
			l.Log("level", "info", "message", "kvm crd has not been deleted")
		}

		certList, err := h.G8sClient().CoreV1alpha1().CertConfigs(v1.NamespaceDefault).List(v1.ListOptions{
			LabelSelector: crdLabelSelector(h.TargetNamespace()),
		})
		if err != nil {
			return microerror.Mask(err)
		}
		if len(certList.Items) == 0 {
			// resource doesnt exist we are good to continue
			certDeleted = true
			l.Log("level", "info", "message", "cert crd was deleted")
		} else {
			l.Log("level", "info", "message", "cert crd has not been deleted")
		}

		if kvmDeleted && certDeleted {
			// crd resources are gone, we can exit
			return nil
		} else {
			return resourceNotDeleted
		}
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

	return nil
}
