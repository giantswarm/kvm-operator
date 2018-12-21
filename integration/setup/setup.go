// +build k8srequired

package setup

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/kvm-operator/integration/env"
)

const (
	kvmResourceValuesFile     = "/tmp/kvm-operator-values.yaml"
	flannelResourceValuesFile = "/tmp/flannel-operator-values.yaml"
)

// WrapTestMain setup and teardown e2e testing environment.
func WrapTestMain(m *testing.M, config Config) {
	var r int

	ctx := context.Background()

	err := Setup(ctx, config)
	if err != nil {
		config.Logger.Log("level", "error", "message", "setup stage failed", "stack", fmt.Sprintf("%#v", err))
		r = 1
	} else {
		config.Logger.Log("level", "info", "message", "finished setup stage")
		r = m.Run()
		if r != 0 {
			config.Logger.Log("level", "error", "message", "test stage failed")
		}
	}

	if env.KeepResources() != "true" {
		config.Logger.Log("level", "info", "message", "removing all resources")
		err = Teardown(config)
		if err != nil {
			config.Logger.Log("level", "error", "message", "teardown stage failed", "stack", fmt.Sprintf("%#v", err))
		}
	} else {
		config.Logger.Log("level", "info", "message", "not removing resources because  env 'KEEP_RESOURCES' is set to true")
	}

	os.Exit(r)
}

// Setup e2e testing environment.
func Setup(ctx context.Context, config Config) error {
	var err error

	err = common(ctx, config)
	if err != nil {
		return microerror.Mask(err)
	}

	err = provider(ctx, config)
	if err != nil {
		return microerror.Mask(err)
	}

	err = config.Guest.Setup(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
