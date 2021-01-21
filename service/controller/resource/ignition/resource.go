package ignition

import (
	"github.com/giantswarm/certs/v3/pkg/certs"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/randomkeys/v2"
)

const (
	// Name is the identifier of the resource.
	Name = "ignition"

	KeyUserData = "user_data"
)

// Config represents the configuration used to create a new config map resource.
type Config struct {
	// Dependencies.
	CertsSearcher certs.Interface
	K8sClient     k8sclient.Interface
	KeyWatcher    randomkeys.Interface
	Logger        micrologger.Logger

	DNSServers                string
	IgnitionPath              string
	NTPServers                string
	SSOPublicKey              string
	DockerhubToken            string
	RegistryDomain            string
	RegistryMirrors           []string
	DockerDaemonCIDR          string
	ImagePullProgressDeadline string
	NetworkSetupDockerImage   string
	PodInfraContainerImage    string
	SSHUserList               string
}

// Resource implements the config map resource.
type Resource struct {
	// Dependencies.
	certsSearcher certs.Interface
	k8sClient     k8sclient.Interface
	keyWatcher    randomkeys.Interface
	logger        micrologger.Logger

	dnsServers                string
	ignitionPath              string
	ntpServers                string
	ssoPublicKey              string
	dockerhubToken            string
	registryDomain            string
	registryMirrors           []string
	dockerDaemonCIDR          string
	imagePullProgressDeadline string
	networkSetupDockerImage   string
	podInfraContainerImage    string
	sshUserList               string
}

// New creates a new configured config map resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.CertsSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CertsSearcher must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.KeyWatcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.KeyWatcher must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.RegistryDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RegistryDomain must not be empty", config)
	}
	if config.DockerhubToken == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.DockerhubToken must not be empty", config)
	}
	if config.IgnitionPath == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.IgnitionPath must not be empty", config)
	}

	newService := &Resource{
		// Dependencies.
		certsSearcher:   config.CertsSearcher,
		k8sClient:       config.K8sClient,
		keyWatcher:      config.KeyWatcher,
		logger:          config.Logger,
		registryDomain:  config.RegistryDomain,
		ignitionPath:    config.IgnitionPath,
		dockerhubToken:  config.DockerhubToken,
		registryMirrors: config.RegistryMirrors,
	}

	return newService, nil
}

func (r *Resource) Name() string {
	return Name
}
