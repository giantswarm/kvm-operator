// +build k8srequired

package setup

import (
	"fmt"
)

func clusterRole(clusterID, operator string) string {
	return fmt.Sprintf("%s-%s", clusterID, operator)
}

func clusterRolePSP(clusterID, operator string) string {
	return fmt.Sprintf("%s-%s-psp", clusterID, operator)
}
