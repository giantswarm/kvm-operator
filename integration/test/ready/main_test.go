// +build k8srequired

package ready

import (
	"os"
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/kvm-operator/integration/env"
	"github.com/giantswarm/kvm-operator/integration/setup"
)

var (
	g *framework.Guest
	h *framework.Host
)

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	var err error

	var logger micrologger.Logger
	{
		c := micrologger.Config{
			IOWriter: os.Stdout,
		}
		logger, err = micrologger.New(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := framework.GuestConfig{
			Logger:       logger,
			ClusterID:    env.ClusterID(),
			CommonDomain: env.CommonDomain(),
		}
		g, err = framework.NewGuest(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := framework.HostConfig{
			Logger:          logger,
			ClusterID:       env.ClusterID(),
			TargetNamespace: env.ClusterID(),
			VaultToken:      env.VaultToken(),
		}
		h, err = framework.NewHost(c)
		if err != nil {
			panic(err.Error())
		}
	}

	setup.WrapTestMain(g, h, m)
}
