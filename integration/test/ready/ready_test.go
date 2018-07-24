// +build k8srequired

package ready

import (
	"testing"
)

// TestGuestReady ensure a guest cluster is up and ready.
//
// The hard job here is done by e2e testing framework:
// 1. setup a host cluster environment
// 2. run kvm-operator in a host cluster environment
// 3. trigger guest cluster creation
// 4. wait until guest cluster is ready
// 5. hand over execution to this test
func TestGuestReady(t *testing.T) {
	t.Log("Guest cluster is ready")
}
