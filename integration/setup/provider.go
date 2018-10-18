// +build k8srequired

package setup

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	gotemplate "text/template"

	cenkaltibackoff "github.com/cenkalti/backoff"
	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
	"github.com/giantswarm/microstorage/retrystorage"

	"github.com/giantswarm/kvm-operator/integration/env"
	"github.com/giantswarm/kvm-operator/integration/ipam"
	"github.com/giantswarm/kvm-operator/integration/key"
	"github.com/giantswarm/kvm-operator/integration/rangepool"
	"github.com/giantswarm/kvm-operator/integration/storage"
	"github.com/giantswarm/kvm-operator/integration/template"
)

// provider installs the operator and tenant cluster CR.
func provider(config Config) error {
	ctx := context.Background()

	{
		c := chartvalues.FlannelOperatorConfig{
			ClusterName: env.ClusterID(),
			ClusterRole: chartvalues.FlannelOperatorClusterRole{
				BindingName: key.ClusterRole("flannel-operator"),
				Name:        key.ClusterRole("flannel-operator"),
			},
			ClusterRolePSP: chartvalues.FlannelOperatorClusterRole{
				BindingName: key.ClusterRolePSP("flannel-operator"),
				Name:        key.ClusterRolePSP("flannel-operator"),
			},
			Namespace: env.TargetNamespace(),
			PSP: chartvalues.FlannelOperatorPSP{
				Name: key.PSPName("flannel-operator"),
			},
			RegistryPullSecret: env.RegistryPullSecret(),
		}

		values, err := chartvalues.NewFlannelOperator(c)
		if err != nil {
			return microerror.Mask(err)
		}

		err = config.Release.InstallOperator(ctx, key.ReleaseName("flannel-operator"), release.NewStableVersion(), values, corev1alpha1.NewFlannelConfigCRD())
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		c := chartvalues.KVMOperatorConfig{
			ClusterName: env.ClusterID(),
			ClusterRole: chartvalues.KVMOperatorClusterRole{
				BindingName: key.ClusterRole("kvm-operator"),
				Name:        key.ClusterRole("kvm-operator"),
			},
			ClusterRolePSP: chartvalues.KVMOperatorClusterRole{
				BindingName: key.ClusterRolePSP("kvm-operator"),
				Name:        key.ClusterRolePSP("kvm-operator"),
			},
			Namespace: env.TargetNamespace(),
			PSP: chartvalues.KVMOperatorPSP{
				Name: key.PSPName("kvm-operator"),
			},
			RegistryPullSecret: env.RegistryPullSecret(),
		}

		values, err := chartvalues.NewKVMOperator(c)
		if err != nil {
			return microerror.Mask(err)
		}

		err = config.Release.InstallOperator(context.Background(), key.ReleaseName("kvm-operator"), release.NewVersion(env.CircleSHA()), values, providerv1alpha1.NewKVMConfigCRD())
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err := installKVMResource(config.Host)
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

		b := func() cenkaltibackoff.BackOff {
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
