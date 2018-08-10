// +build k8srequired

package teardown

import (
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"fmt"
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

	{
		err = framework.HelmCmd(fmt.Sprintf("delete %s-cert-config-e2e --purge", h.TargetNamespace()))
		if err != nil {
			return microerror.Mask(err)
		}
		err = framework.HelmCmd(fmt.Sprintf("delete %s-aws-config-e2e --purge", h.TargetNamespace()))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
