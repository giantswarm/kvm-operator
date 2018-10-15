package chartvalues

import (
	"github.com/giantswarm/e2etemplates/internal/render"
	"github.com/giantswarm/microerror"
)

type CertOperatorConfig struct {
	ClusterName        string
	ClusterRole        CertOperatorClusterRole
	ClusterRolePSP     CertOperatorClusterRole
	CommonDomain       string
	RegistryPullSecret string
	PSP                CertOperatorPSP
	Vault              CertOperatorVault
}

type CertOperatorClusterRole struct {
	BindingName string
	Name        string
}

type CertOperatorPSP struct {
	Name string
}

type CertOperatorVault struct {
	Token string
}

func NewCertOperator(config CertOperatorConfig) (string, error) {
	if config.ClusterName == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.ClusterName must not be empty", config)
	}
	if config.ClusterRole.BindingName == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.ClusterRole.BindingName must not be empty", config)
	}
	if config.ClusterRole.Name == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.ClusterRole.Name must not be empty", config)
	}
	if config.ClusterRolePSP.BindingName == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.ClusterRolePSP.BindingName must not be empty", config)
	}
	if config.ClusterRolePSP.Name == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.ClusterRolePSP.Name must not be empty", config)
	}
	if config.CommonDomain == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.CommonDomain must not be empty", config)
	}
	if config.PSP.Name == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.PSP.Name must not be empty", config)
	}
	if config.RegistryPullSecret == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.RegistryPullSecret must not be empty", config)
	}
	if config.Vault.Token == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Vault.Token must not be empty", config)
	}

	values, err := render.Render(certOperatorTemplate, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return values, nil
}
