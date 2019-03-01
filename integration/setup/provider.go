// +build k8srequired

package setup

import (
	"context"

	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/kvm-operator/integration/env"
	"github.com/giantswarm/kvm-operator/integration/ipam"
	"github.com/giantswarm/kvm-operator/integration/key"
	"github.com/giantswarm/kvm-operator/integration/rangepool"
)

// provider installs the operator and tenant cluster CR.
func provider(ctx context.Context, config Config) error {
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

		err = config.Release.InstallOperator(ctx, "flannel-operator", release.NewStableVersion(), values, corev1alpha1.NewFlannelConfigCRD())
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

		err = config.Release.InstallOperator(context.Background(), "kvm-operator", release.NewVersion(env.CircleSHA()), values, providerv1alpha1.NewKVMConfigCRD())
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err := installKVMResource(config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func installKVMResource(config Config) error {
	ctx := context.Background()

	var httpPort, httpsPort, vni int
	{
		rangePool, err := rangepool.InitRangePool(config.Storage, config.Logger)
		if err != nil {
			return microerror.Mask(err)
		}

		{
			vni, err = rangepool.GenerateVNI(ctx, rangePool, env.ClusterID())
			if err != nil {
				return microerror.Mask(err)
			}
		}

		{
			httpPort, httpsPort, err = rangepool.GenerateIngressNodePorts(ctx, rangePool, env.ClusterID())
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	{
		network, err := ipam.GenerateFlannelNetwork(ctx, env.ClusterID(), config.Storage, config.Logger)
		if err != nil {
			return microerror.Mask(err)
		}

		c := chartvalues.APIExtensionsFlannelConfigE2EConfig{
			ClusterID: env.ClusterID(),
			Network:   network,
			VNI:       vni,
		}

		values, err := chartvalues.NewAPIExtensionsFlannelConfigE2E(c)
		if err != nil {
			return microerror.Mask(err)
		}

		err = config.Release.EnsureInstalled(ctx, key.FlannelReleaseName(env.TargetNamespace()), release.NewStableChartInfo("apiextensions-flannel-config-e2e-chart"), values)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		c := chartvalues.APIExtensionsKVMConfigE2EConfig{
			ClusterID:            env.ClusterID(),
			HttpNodePort:         httpPort,
			HttpsNodePort:        httpsPort,
			VersionBundleVersion: env.VersionBundleVersion(),
			VNI:                  vni,
		}

		values, err := chartvalues.NewAPIExtensionsKVMConfigE2E(c)
		if err != nil {
			return microerror.Mask(err)
		}

		err = config.Release.EnsureInstalled(ctx, key.KVMReleaseName(env.TargetNamespace()), release.NewStableChartInfo("apiextensions-kvm-config-e2e-chart"), values)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
