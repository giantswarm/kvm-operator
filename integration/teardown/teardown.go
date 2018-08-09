// +build k8srequired

package teardown

import (
	"fmt"

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
		err = framework.HelmCmd(fmt.Sprintf("delete kvm-operator --namespace %s --purge", targetNamespace))
		if err != nil {
			return microerror.Mask(err)
		}
		err = framework.HelmCmd(fmt.Sprintf("delete cert-operator --namespace %s --purge", targetNamespace))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = framework.HelmCmd(fmt.Sprintf("delete cert-config-e2e --namespace %s --purge", targetNamespace))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = h.K8sClient().CoreV1().Namespaces().Delete(targetNamespace, &metav1.DeleteOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
