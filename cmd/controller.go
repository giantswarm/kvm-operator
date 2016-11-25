package cmd

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/cluster-controller/controller"
)

var (
	inCluster bool

	apiServer    string
	caFilePath   string
	certFilePath string
	keyFilePath  string

	listenAddress string
)

func init() {
	RootCmd.AddCommand(controllerCmd)

	controllerCmd.Flags().BoolVar(&inCluster, "in-cluster", true, "Is the controller running in a cluster")

	controllerCmd.Flags().StringVar(&apiServer, "api-server", "", "API server URL")
	controllerCmd.Flags().StringVar(&caFilePath, "ca-file-path", "", "CA file path")
	controllerCmd.Flags().StringVar(&certFilePath, "cert-file-path", "", "Cert file path")
	controllerCmd.Flags().StringVar(&keyFilePath, "key-file-path", "", "Key file path")

	controllerCmd.Flags().StringVar(&listenAddress, "listen-address", "127.0.0.1:8000", "Listen address for server")
}

var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Start the controller",
	Run:   controllerRun,
}

func controllerRun(cmd *cobra.Command, args []string) {
	config := controller.Config{
		InCluster: inCluster,

		APIServer:    apiServer,
		CAFilePath:   caFilePath,
		CertFilePath: certFilePath,
		KeyFilePath:  keyFilePath,

		ListenAddress: listenAddress,
	}

	controller := controller.New(config)
	controller.Start()
}
