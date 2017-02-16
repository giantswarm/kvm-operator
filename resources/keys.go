package resources

import (
	"fmt"
)

func clusterDomain(sub, clusterID, domain string) string {
	return fmt.Sprintf("%s.%s.%s", sub, clusterID, domain)
}

func networkBridgeName(ID string) string {
	return fmt.Sprintf("br-%s", ID)
}

func networkEnvFilePath(ID string) string {
	return fmt.Sprintf("/run/flannel/networks/%s.env", networkBridgeName(ID))
}
