// +build k8srequired

package teardown

import (
	"fmt"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
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
		_, err := h.G8sClient().ProviderV1alpha1().KVMConfigs(v1.NamespaceDefault).Get(h.TargetNamespace())
		if err != nil {
			// resource doesnt exist we are good to continue
			kvmDeleted = true
		}

		_, err := h.G8sClient().CoreV1alpha1().CertConfigs(v1.NamespaceDefault).Get(h.TargetNamespace())
		if err != nil {
			// resource doesnt exist we are good to continue
			certDeleted = true
		}

		if kvmDeleted && certDeleted {
			return nil
		} else {
			return microerror.New("resource still exist")
		}
	}
	b := backoff.NewExponential(framework.ShortMaxWait, framework.ShortMaxInterval)
	n := backoff.NewNotifier(l, context.Background())
	err = backoff.RetryNotify(o, b, n)

	{
		err = h.K8sClient().CoreV1().Namespaces().Delete(h.TargetNamespace(), &metav1.DeleteOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
