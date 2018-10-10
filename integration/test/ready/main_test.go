// +build k8srequired

package ready

import (
	"testing"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2e-harness/pkg/framework/filelogger"
	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/kvm-operator/integration/env"
	"github.com/giantswarm/kvm-operator/integration/setup"
)

var (
	g *framework.Guest
	h *framework.Host
	r *release.Release
)

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	var err error

	var logger micrologger.Logger
	{
		c := micrologger.Config{}

		logger, err = micrologger.New(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := framework.GuestConfig{
			Logger: logger,

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
			Logger: logger,

			ClusterID:       env.ClusterID(),
			TargetNamespace: env.TargetNamespace(),
			VaultToken:      env.VaultToken(),
		}
		h, err = framework.NewHost(c)
		if err != nil {
			panic(err.Error())
		}
	}

	var fileLogger *filelogger.FileLogger
	{
		fc := filelogger.Config{
			Backoff:   backoff.NewExponential(backoff.ShortMaxWait, backoff.LongMaxInterval),
			K8sClient: h.K8sClient(),
			Logger:    logger,
		}
		fileLogger, err = filelogger.New(fc)
		if err != nil {
			panic(err.Error())
		}
	}
	var helmClient *helmclient.Client
	{
		c := helmclient.Config{
			Logger:          logger,
			K8sClient:       h.K8sClient(),
			RestConfig:      h.RestConfig(),
			TillerNamespace: "kube-system",
		}
		helmClient, err = helmclient.New(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := release.Config{
			ExtClient:  h.ExtClient(),
			FileLogger: fileLogger,
			G8sClient:  h.G8sClient(),
			HelmClient: helmClient,
			K8sClient:  h.K8sClient(),
			Logger:     logger,

			Namespace: h.TargetNamespace(),
		}
		r, err = release.New(c)
		if err != nil {
			panic(err.Error())
		}
	}

	setup.WrapTestMain(g, h, helmClient, m, r, logger)
}
