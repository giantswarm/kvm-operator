package resources

import (
	"fmt"
)

func bridgeIPConfigmapName(node string) string {
	return fmt.Sprintf("bridge-ip-configmap-%s", node)
}

func bridgeIPConfigmapPath(node string) string {
	return fmt.Sprintf("/tmp/%s.json", bridgeIPConfigmapName(node))
}

func clusterDomain(sub, clusterID, domain string) string {
	return fmt.Sprintf("%s.%s.%s", sub, clusterID, domain)
}

func networkBridgeName(ID string) string {
	return fmt.Sprintf("br-%s", ID)
}

func networkEnvFilePath(ID string) string {
	return fmt.Sprintf("/run/flannel/networks/%s.env", networkBridgeName(ID))
}
