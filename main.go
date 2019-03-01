package main

import (
	"fmt"

	"github.com/giantswarm/kvm-operator/flag"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/microkit/command"
	microserver "github.com/giantswarm/microkit/server"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/viper"

	"github.com/giantswarm/kvm-operator/server"
	"github.com/giantswarm/kvm-operator/service"
)

var (
	description string     = "The kvm-operator handles Kubernetes clusters running on a Kubernetes cluster."
	f           *flag.Flag = flag.New()
	gitCommit   string     = "n/a"
	name        string     = "kvm-operator"
	source      string     = "https://github.com/giantswarm/kvm-operator"
)

func main() {
	err := mainError()
	if err != nil {
		panic(fmt.Sprintf("%#v\n", err))
	}
}

func mainError() error {
	var err error

	// Create a new logger which is used by all packages.
	var newLogger micrologger.Logger
	{
		c := micrologger.Config{}

		newLogger, err = micrologger.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// We define a server factory to create the custom server once all command
	// line flags are parsed and all microservice configuration is storted out.
	newServerFactory := func(v *viper.Viper) microserver.Server {
		// Create a new custom service which implements business logic.
		var newService *service.Service
		{
			c := service.Config{
				Logger: newLogger,

				Description: description,
				Flag:        f,
				GitCommit:   gitCommit,
				Name:        name,
				Source:      source,
				Viper:       v,
			}

			newService, err = service.New(c)
			if err != nil {
				panic(fmt.Sprintf("%#v", err))
			}
			go newService.Boot()
		}

		// Create a new custom server which bundles our endpoints.
		var newServer microserver.Server
		{
			c := server.Config{
				Logger:  newLogger,
				Service: newService,
				Viper:   v,

				ProjectName: name,
			}

			newServer, err = server.New(c)
			if err != nil {
				panic(fmt.Sprintf("%#v", err))
			}
		}

		return newServer
	}

	// Create a new microkit command which manages our custom microservice.
	var newCommand command.Command
	{
		c := command.Config{
			Logger:        newLogger,
			ServerFactory: newServerFactory,

			Description:    description,
			GitCommit:      gitCommit,
			Name:           name,
			Source:         source,
			VersionBundles: service.NewVersionBundles(),
		}

		newCommand, err = command.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	daemonCommand := newCommand.DaemonCommand().CobraCommand()

	daemonCommand.PersistentFlags().String(f.Service.Installation.DNS.Servers, "", "Comma separated list of DNS servers.")

	daemonCommand.PersistentFlags().String(f.Service.Installation.Tenant.Kubernetes.API.Auth.Provider.OIDC.ClientID, "", "OIDC authorization provider ClientID.")
	daemonCommand.PersistentFlags().String(f.Service.Installation.Tenant.Kubernetes.API.Auth.Provider.OIDC.IssuerURL, "", "OIDC authorization provider IssuerURL.")
	daemonCommand.PersistentFlags().String(f.Service.Installation.Tenant.Kubernetes.API.Auth.Provider.OIDC.UsernameClaim, "", "OIDC authorization provider UsernameClaim.")
	daemonCommand.PersistentFlags().String(f.Service.Installation.Tenant.Kubernetes.API.Auth.Provider.OIDC.GroupsClaim, "", "OIDC authorization provider GroupsClaim.")

	daemonCommand.PersistentFlags().String(f.Service.CRD.LabelSelector, "", "Label selector for CRD informer ListOptions.")

	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.Address, "http://127.0.0.1:6443", "Address used to connect to Kubernetes. When empty in-cluster config is created.")
	daemonCommand.PersistentFlags().Bool(f.Service.Kubernetes.InCluster, false, "Whether to use the in-cluster config to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CAFile, "", "Certificate authority file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CrtFile, "", "Certificate file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.KeyFile, "", "Key file path to use to authenticate with Kubernetes.")

	daemonCommand.PersistentFlags().String(f.Service.Tenant.Ignition.Path, "/opt/ignition", "Default path for the ignition base directory.")
	daemonCommand.PersistentFlags().String(f.Service.Tenant.SSH.SSOPublicKey, "", "Public key for trusted SSO CA.")
	daemonCommand.PersistentFlags().Bool(f.Service.Tenant.Update.Enabled, false, "Whether updates of tenant cluster nodes are allowed to be processed upon reconciliation.")

	newCommand.CobraCommand().Execute()

	return nil
}
