package cloudconfig

import (
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	FileOwnerUserName  = "root"
	FileOwnerGroupName = "root"
	FilePermission     = 0700

	IscsiInitiatorNameFilePath    = "/etc/iscsi/initiatorname.iscsi"
	IscsiInitiatorFilePermissions = 0644

	IscsiConfigFilePath        = "/etc/iscsi/iscsid.conf"
	IscsiConfigFilePermissions = 0644
	IscsiConfigFileContent     = `
# Check for active mounts on devices reachable through a session
# and refuse to logout if there are any.  Defaults to "No".
iscsid.safe_logout = Yes`
)

// Config represents the configuration used to create a cloud config service.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger

	IgnitionPath string
	OIDC         OIDCConfig
	SSOPublicKey string
}

// DefaultConfig provides a default configuration to create a new cloud config
// service by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger: nil,
	}
}

// CloudConfig implements the cloud config service interface.
type CloudConfig struct {
	// Dependencies.
	logger micrologger.Logger

	ignitionPath    string
	k8sAPIExtraArgs []string
	ssoPublicKey    string
}

// OIDCConfig represents the configuration of the OIDC authorization provider
type OIDCConfig struct {
	ClientID      string
	IssuerURL     string
	UsernameClaim string
	GroupsClaim   string
}

// New creates a new configured cloud config service.
func New(config Config) (*CloudConfig, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	if config.IgnitionPath == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.IgnitionPath must not be empty", config)
	}

	var k8sAPIExtraArgs []string
	{
		if config.OIDC.ClientID != "" {
			k8sAPIExtraArgs = append(k8sAPIExtraArgs, fmt.Sprintf("--oidc-client-id=%s", config.OIDC.ClientID))
		}
		if config.OIDC.IssuerURL != "" {
			k8sAPIExtraArgs = append(k8sAPIExtraArgs, fmt.Sprintf("--oidc-issuer-url=%s", config.OIDC.IssuerURL))
		}
		if config.OIDC.UsernameClaim != "" {
			k8sAPIExtraArgs = append(k8sAPIExtraArgs, fmt.Sprintf("--oidc-username-claim=%s", config.OIDC.UsernameClaim))
		}
		if config.OIDC.GroupsClaim != "" {
			k8sAPIExtraArgs = append(k8sAPIExtraArgs, fmt.Sprintf("--oidc-groups-claim=%s", config.OIDC.GroupsClaim))
		}
	}

	newCloudConfig := &CloudConfig{
		// Dependencies.
		logger: config.Logger,

		ignitionPath:    config.IgnitionPath,
		k8sAPIExtraArgs: k8sAPIExtraArgs,
		ssoPublicKey:    config.SSOPublicKey,
	}

	return newCloudConfig, nil
}
