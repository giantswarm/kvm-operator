// +build k8srequired

package setup

import (
	"context"
	"fmt"

	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/kvm-operator/integration/env"
	"github.com/giantswarm/kvm-operator/integration/key"
)

// common installs components required to run the operator.
func common(config Config) error {
	ctx := context.Background()

	{
		c := chartvalues.CertOperatorConfig{
			ClusterRole: chartvalues.CertOperatorConfigClusterRole{
				BindingName: key.ClusterRole("cert-operator"),
				Name:        key.ClusterRole("cert-operator"),
			},
			ClusterRolePSP: chartvalues.CertOperatorConfigClusterRole{
				BindingName: key.ClusterRolePSP("cert-operator"),
				Name:        key.ClusterRolePSP("cert-operator"),
			},
			CommonDomain: env.CommonDomain(),
			CRD: chartvalues.CertOperatorConfigCRD{
				LabelSelector: key.LabelSelector(),
			},
			Namespace: env.TargetNamespace(),
			PSP: chartvalues.CertOperatorPSP{
				Name: key.PSPName("cert-operator"),
			},
			RegistryPullSecret: env.RegistryPullSecret(),
			Vault: chartvalues.CertOperatorVault{
				Token: env.VaultToken(),
			},
		}

		values, err := chartvalues.NewCertOperator(c)
		if err != nil {
			return microerror.Mask(err)
		}

		err = config.Release.InstallOperator(ctx, "cert-operator", release.NewStableVersion(), values, corev1alpha1.NewCertConfigCRD())
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		c := chartvalues.NodeOperatorConfig{
			RegistryPullSecret: env.RegistryPullSecret(),
		}

		values, err := chartvalues.NewNodeOperator(c)
		if err != nil {
			return microerror.Mask(err)
		}

		err = config.Release.InstallOperator(ctx, "node-operator", release.NewStableVersion(), values, corev1alpha1.NewNodeConfigCRD())
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		c := chartvalues.E2ESetupCertsConfig{
			Cluster: chartvalues.E2ESetupCertsConfigCluster{
				ID: env.ClusterID(),
			},
			CommonDomain: env.CommonDomain(),
		}

		values, err := chartvalues.NewE2ESetupCerts(c)
		if err != nil {
			return microerror.Mask(err)
		}

		err = config.Release.Install(ctx, fmt.Sprintf("e2esetup-certs-%s", env.ClusterID()), release.NewStableChartInfo("e2esetup-certs-chart"), values, config.Release.Condition().SecretExists(ctx, "default", fmt.Sprintf("%s-api", env.ClusterID())))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
