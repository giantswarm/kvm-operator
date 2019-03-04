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

func FlannelReleaseName(namespace string) string {
	return fmt.Sprintf("%s-flannel-config-e2e", namespace)
}

func KVMReleaseName(namespace string) string {
	return fmt.Sprintf("%s-kvm-config-e2e", namespace)
}

func LabelSelector() string {
	return fmt.Sprintf("giantswarm.io/cluster=%s", env.ClusterID())
}

func PSPName(operator string) string {
	return fmt.Sprintf("%s-%s", env.ClusterID(), operator)
}
