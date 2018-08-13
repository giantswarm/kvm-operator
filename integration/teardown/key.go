package teardown

import "fmt"

func crdLabelSelector(clusterID string) string {
	return fmt.Sprintf("clusterID=%s", clusterID)
}
