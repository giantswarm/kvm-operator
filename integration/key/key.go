// +build k8srequired

package key

import (
	"fmt"

	"github.com/giantswarm/kvm-operator/integration/env"
)

func ClusterRole(operator string) string {
	return fmt.Sprintf("%s-%s", env.ClusterID(), operator)
}

func ClusterRolePSP(operator string) string {
	return fmt.Sprintf("%s-%s-psp", env.ClusterID(), operator)
}

func ReleaseName(operator string) string {
	return fmt.Sprintf("%s-%s", env.TargetNamespace(), operator)
}

func PSPName(operator string) string {
	return fmt.Sprintf("%s-%s", env.ClusterID(), operator)
}
