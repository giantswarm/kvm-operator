// +build k8srequired

package setup

import (
	"fmt"
	"strings"
)

func crdLabelSelector(clusterID string) string {
	return fmt.Sprintf("clusterID=%s", clusterID)
}

func getClusterID(targetNamespace string) string {
	return strings.TrimRight(targetNamespace, "-op")
}
