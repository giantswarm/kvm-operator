// +build k8srequired

package setup

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
	gotemplate "text/template"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"

	"github.com/giantswarm/kvm-operator/integration/env"
	"github.com/giantswarm/kvm-operator/integration/ipam"
	"github.com/giantswarm/kvm-operator/integration/rangepool"
	"github.com/giantswarm/kvm-operator/integration/storage"
	"github.com/giantswarm/kvm-operator/integration/teardown"
	"github.com/giantswarm/kvm-operator/integration/template"
)

const (
	kvmResourceValuesFile     = "/tmp/kvm-operator-values.yaml"
	flannelResourceValuesFile = "/tmp/flannel-operator-values.yaml"
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
		fmt.Printf("\n Removing all resources.\n")
		teardown.Teardown(g, h)
	} else {
		fmt.Printf("\nNot removing resources becasue  env 'KEEP_RESOURCES' is set to true.\n")
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

	err = g.Setup()
	if err != nil {
		return microerror.Mask(err)
	}

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
		err = h.InstallStableOperator("flannel-operator", "flannelconfig", template.FlannelOperatorChartValues)
		if err != nil {
			return microerror.Mask(err)
		}
		err = h.InstallStableOperator("node-operator", "drainerconfig", e2etemplates.NodeOperatorChartValues)
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

		err = installKVMResource(h)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func installKVMResource(h *framework.Host) error {
	var err error
	ctx := context.Background()

	var l micrologger.Logger
	{
		c := micrologger.Config{}

		l, err = micrologger.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var crdStorage microstorage.Storage
	{
		crdStorage, err = storage.InitCRDStorage(ctx, h, l)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var kvmResourceChartValues template.KVMConfigE2eChartValues
	{
		kvmResourceChartValues.ClusterID = env.ClusterID()

		rangePool, err := rangepool.InitRangePool(crdStorage, l)
		if err != nil {
			return microerror.Mask(err)
		}

		{
			vni, err := rangepool.GenerateVNI(ctx, rangePool, env.ClusterID())
			if err != nil {
				return microerror.Mask(err)
			}
			kvmResourceChartValues.VNI = vni
		}

		{
			httpPort, httpsPort, err := rangepool.GenerateIngressNodePorts(ctx, rangePool, env.ClusterID())
			if err != nil {
				return microerror.Mask(err)
			}
			kvmResourceChartValues.HttpNodePort = httpPort
			kvmResourceChartValues.HttpsNodePort = httpsPort
		}

		kvmResourceChartValues.VersionBundleVersion = env.VersionBundleVersion()
	}

	var flannelResourceChartValues template.FlannelConfigE2eChartValues
	{
		flannelResourceChartValues.ClusterID = env.ClusterID()
		flannelResourceChartValues.VNI = kvmResourceChartValues.VNI

		network, err := ipam.GenerateFlannelNetwork(ctx, env.ClusterID(), crdStorage, l)
		if err != nil {
			return microerror.Mask(err)
		}
		flannelResourceChartValues.Network = network
	}

	o := func() error {
		// NOTE we ignore errors here because we cannot get really useful error
		// handling done. This here should anyway only be a quick fix until we use
		// the helm client lib. Then error handling will be better.
		framework.HelmCmd(fmt.Sprintf("delete --purge %s-flannel-config-e2e", h.TargetNamespace()))

		var buffer bytes.Buffer

		tmpl, err := gotemplate.New("flannel-e2e-values").Parse(template.ApiextensionsFlannelConfigE2EChartValues)
		if err != nil {
			return microerror.Mask(err)
		}

		err = tmpl.Execute(&buffer, flannelResourceChartValues)

		if err != nil {
			return microerror.Mask(err)
		}

		err = ioutil.WriteFile(flannelResourceValuesFile, buffer.Bytes(), 0644)
		if err != nil {
			return microerror.Mask(err)
		}

		err = framework.HelmCmd(fmt.Sprintf("registry install quay.io/giantswarm/apiextensions-flannel-config-e2e-chart:stable -- -n %s-flannel-config-e2e --values %s", h.TargetNamespace(), flannelResourceValuesFile))
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}
	b := backoff.NewExponential(framework.ShortMaxWait, framework.ShortMaxInterval)
	n := backoff.NewNotifier(l, context.Background())
	err = backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	o = func() error {
		// NOTE we ignore errors here because we cannot get really useful error
		// handling done. This here should anyway only be a quick fix until we use
		// the helm client lib. Then error handling will be better.
		framework.HelmCmd(fmt.Sprintf("delete --purge %s-kvm-config-e2e", h.TargetNamespace()))

		var buffer bytes.Buffer

		tmpl, err := gotemplate.New("kvm-e2e-values").Parse(template.ApiextensionsKVMConfigE2EChartValues)
		if err != nil {
			return microerror.Mask(err)
		}

		err = tmpl.Execute(&buffer, kvmResourceChartValues)

		if err != nil {
			return microerror.Mask(err)
		}

		err = ioutil.WriteFile(kvmResourceValuesFile, buffer.Bytes(), 0644)
		if err != nil {
			return microerror.Mask(err)
		}

		err = framework.HelmCmd(fmt.Sprintf("registry install quay.io/giantswarm/apiextensions-kvm-config-e2e-chart:stable -- -n %s-kvm-config-e2e --values %s", h.TargetNamespace(), kvmResourceValuesFile))
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}
	b = backoff.NewExponential(framework.ShortMaxWait, framework.ShortMaxInterval)
	n = backoff.NewNotifier(l, context.Background())
	err = backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
