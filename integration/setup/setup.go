// +build k8srequired

package setup

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	gotemplate "text/template"

	cenkalti "github.com/cenkalti/backoff"
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
	"github.com/giantswarm/microstorage/retrystorage"

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
func WrapTestMain(g *framework.Guest, h *framework.Host, m *testing.M, r *release.Release, l micrologger.Logger) {
	var exitCode int

	err := Setup(g, h, r)
	if err != nil {
		l.Log("level", "error", "message", "setup stage failed", "stack", fmt.Sprintf("%#v", err))
		exitCode = 1
	} else {
		l.Log("level", "info", "message", "finished setup stage")
		exitCode = m.Run()
		if exitCode != 0 {
			l.Log("level", "error", "message", "test stage failed")
		}
	}

	if env.KeepResources() != "true" {
		l.Log("level", "info", "message", "removing all resources")
		err = teardown.Teardown(g, h)
		if err != nil {
			l.Log("level", "error", "message", "teardown stage failed", "stack", fmt.Sprintf("%#v", err))

		}
	} else {
		l.Log("level", "info", "message", "not removing resources because  env 'KEEP_RESOURCES' is set to true")
	}

	os.Exit(exitCode)
}

// Setup e2e testing environment.
func Setup(g *framework.Guest, h *framework.Host, r *release.Release) error {
	var err error

	err = Resources(g, h, r)
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
func Resources(g *framework.Guest, h *framework.Host, r *release.Release) error {
	ctx := context.Background()
	var err error

	{
		err = r.InstallOperator(ctx, "cert-operator", release.NewStableVersion(), e2etemplates.CertOperatorChartValues, v1alpha1.NewCertConfigCRD())
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

	var retryingCRDStorage microstorage.Storage
	{
		crdStorage, err := storage.InitCRDStorage(ctx, h, l)
		if err != nil {
			return microerror.Mask(err)
		}

		b := func() cenkalti.BackOff {
			return backoff.NewExponential(backoff.ShortMaxWait, backoff.ShortMaxInterval)
		}

		c := retrystorage.Config{
			Logger:         l,
			Underlying:     crdStorage,
			NewBackOffFunc: b,
		}

		retryingCRDStorage, err = retrystorage.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var kvmResourceChartValues template.KVMConfigE2eChartValues
	{
		kvmResourceChartValues.ClusterID = env.ClusterID()

		rangePool, err := rangepool.InitRangePool(retryingCRDStorage, l)
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

		network, err := ipam.GenerateFlannelNetwork(ctx, env.ClusterID(), retryingCRDStorage, l)
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
	b := backoff.NewExponential(backoff.ShortMaxWait, backoff.ShortMaxInterval)
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
	b = backoff.NewExponential(backoff.ShortMaxWait, backoff.ShortMaxInterval)
	n = backoff.NewNotifier(l, context.Background())
	err = backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
