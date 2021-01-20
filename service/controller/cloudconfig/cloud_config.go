package cloudconfig

import (
	"github.com/giantswarm/certs/v3/pkg/certs"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/randomkeys/v2"
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

const calicoKubeKillScript = `#!/bin/bash

HOSTNAME=$(hostname | tr '[:upper:]' '[:lower:]')

while [ "$(kubectl get nodes $HOSTNAME -o jsonpath='{.metadata.name}')" != "$HOSTNAME" ]; do
  sleep 1
  echo "Waiting for node $HOSTNAME to be registered"
done

sleep 30s

RETRY=5
result=""

while [ "$result" != "ok" ] && [ $RETRY -gt 0 ]; do
  sleep 10s
  echo "Trying to restart k8s services ..."
  let RETRY=$RETRY-1
  kubectl -n kube-system delete pod -l k8s-app=calico-node && \
    sleep 1m && \
    kubectl -n kube-system delete pod -l k8s-app=kube-proxy && \
    kubectl -n kube-system delete pod -l k8s-app=calico-kube-controllers && \
    kubectl -n kube-system delete pod -l k8s-app=coredns && \
    result="ok" || echo "failed"
done

if [ "$result" != "ok" ]; then
  echo "Failed to restart k8s services."
  exit 1
fi

echo "Successfully restarted k8s services."`

// Config represents the configuration used to create a cloud config service.
type Config struct {
	CertsSearcher      certs.Interface
	K8sClient          k8sclient.Interface
	Logger             micrologger.Logger
	RandomKeysSearcher randomkeys.Interface

	OIDC                      OIDCConfig
	APIExtraArgs              []string
	CalicoCIDR                int
	CalicoMTU                 int
	CalicoSubnet              string
	ClusterIPRange            string
	DockerDaemonCIDR          string
	DockerhubToken            string
	ExternalSNAT              bool
	IgnitionPath              string
	ImagePullProgressDeadline string
	KubeletExtraArgs          []string
	ClusterDomain             string
	NetworkSetupDockerImage   string
	PodInfraContainerImage    string
	RegistryDomain            string
	RegistryMirrors           []string
	SSHUserList               string
	SSOPublicKey              string
}

// OIDCConfig represents the configuration of the OIDC authorization provider
type OIDCConfig struct {
	ClientID       string
	IssuerURL      string
	UsernameClaim  string
	UsernamePrefix string
	GroupsClaim    string
	GroupsPrefix   string
}

func (c Config) Validate() error {
	if c.CertsSearcher == nil {
		return microerror.Maskf(invalidConfigError, "%T.CertsSearcher must not be empty", c)
	}
	if c.K8sClient == nil {
		return microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", c)
	}
	if c.Logger == nil {
		return microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", c)
	}
	if c.RandomKeysSearcher == nil {
		return microerror.Maskf(invalidConfigError, "%T.RandomKeysSearcher must not be empty", c)
	}

	if c.CalicoCIDR == 0 {
		return microerror.Maskf(invalidConfigError, "%T.CalicoCIDR must not be empty", c)
	}
	if c.CalicoMTU == 0 {
		return microerror.Maskf(invalidConfigError, "%T.CalicoMTU must not be empty", c)
	}
	if c.CalicoSubnet == "" {
		return microerror.Maskf(invalidConfigError, "%T.CalicoSubnet must not be empty", c)
	}
	if c.ClusterDomain == "" {
		return microerror.Maskf(invalidConfigError, "%T.ClusterDomain must not be empty", c)
	}
	if c.ClusterIPRange == "" {
		return microerror.Maskf(invalidConfigError, "%T.ClusterIPRange must not be empty", c)
	}
	if c.DockerDaemonCIDR == "" {
		return microerror.Maskf(invalidConfigError, "%T.DockerDaemonCIDR must not be empty", c)
	}
	if c.DockerhubToken == "" {
		return microerror.Maskf(invalidConfigError, "%T.DockerhubToken must not be empty", c)
	}

	if c.IgnitionPath == "" {
		return microerror.Maskf(invalidConfigError, "%T.IgnitionPath must not be empty", c)
	}
	if c.ImagePullProgressDeadline == "" {
		return microerror.Maskf(invalidConfigError, "%T.ImagePullProgressDeadline must not be empty", c)
	}
	if c.NetworkSetupDockerImage == "" {
		return microerror.Maskf(invalidConfigError, "%T.NetworkSetupDockerImage must not be empty", c)
	}
	if c.RegistryDomain == "" {
		return microerror.Maskf(invalidConfigError, "%T.RegistryDomain must not be empty", c)
	}
	if c.SSHUserList == "" {
		return microerror.Maskf(invalidConfigError, "%T.SSHUserList must not be empty", c)
	}
	if c.SSOPublicKey == "" {
		return microerror.Maskf(invalidConfigError, "%T.SSOPublicKey must not be empty", c)
	}

	return nil
}
