// +build k8srequired

package teardown

import (
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/integration/env"
)

// Teardown e2e testing environment.
func Teardown(g *framework.Guest, h *framework.Host) error {
	var err error
	targetNamespace := env.ClusterID()

	{
		err = h.K8sClient().CoreV1().Namespaces().Delete(targetNamespace, &metav1.DeleteOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
