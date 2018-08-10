// +build k8srequired

package setup

import (
	"log"
	"os"
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/kvm-operator/integration/env"
	"github.com/giantswarm/kvm-operator/integration/teardown"
	"github.com/giantswarm/kvm-operator/integration/template"
)

const (
	kvmResourceValuesFile = "/tmp/kvm-operator-values.yaml"
)

// WrapTestMain setup and teardown e2e testing environment.
func WrapTestMain(g *framework.Guest, h *framework.Host, m *testing.M) {
	var r int

	err := Setup(g, h)
	if err != nil {
		log.Printf("%#v\n", err)
		r = 1
	} else {
		r = m.Run()
	}

	if env.KeepResources() != "true" {
		teardown.Teardown(g, h)
	}

	os.Exit(r)
}

// Setup e2e testing environment.
func Setup(g *framework.Guest, h *framework.Host) error {
	var err error

	err = Resources(g, h)
	if err != nil {
		return microerror.Mask(err)
	}

	// TODO(r7vme): Enable, when real kvm host cluster will be used.
	//err = g.Setup()
	//if err != nil {
	//	return microerror.Mask(err)
	//}

	return nil
}

// Resources install required charts.
func Resources(g *framework.Guest, h *framework.Host) error {
	var err error

	{
		err = h.InstallStableOperator("cert-operator", "certconfig", e2etemplates.CertOperatorChartValues)
		if err != nil {
			return microerror.Mask(err)
		}

		err = h.InstallBranchOperator("kvm-operator", "kvmconfig", template.KVMOperatorChartValues)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = h.InstallCertResource()
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
