// +build k8srequired

package setup

import (
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/kvm-operator/integration/env"
	"github.com/giantswarm/kvm-operator/integration/key"
)

// common installs components required to run the operator.
func common(config Config) error {
	{
		c := chartvalues.CertOperatorConfig{
			ClusterName: env.ClusterID(),
			ClusterRole: chartvalues.CertOperatorClusterRole{
				BindingName: key.ClusterRole("cert-operator"),
				Name:        key.ClusterRole("cert-operator"),
			},
			ClusterRolePSP: chartvalues.CertOperatorClusterRole{
				BindingName: key.ClusterRolePSP("cert-operator"),
				Name:        key.ClusterRolePSP("cert-operator"),
			},
			CommonDomain: env.CommonDomain(),
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

		err = config.Host.InstallStableOperator("cert-operator", "certconfig", values)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err := config.Host.InstallStableOperator("node-operator", "drainerconfig", e2etemplates.NodeOperatorChartValues)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err := config.Host.InstallCertResource()
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
