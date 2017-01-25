package resources

import (
	"fmt"
)

func networkBridgeName(ID string) string {
	return fmt.Sprintf("br-%s", ID)
}
