package cmd

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/cluster-controller/controller"
)

var (
	kubernetesAPIServer    string
	kubernetesCAFilePath   string
	kubernetesCertFilePath string
	kubernetesKeyFilePath  string

	listenAddress string
)

func init() {
	RootCmd.AddCommand(controllerCmd)

	controllerCmd.Flags().StringVar(&kubernetesAPIServer, "kubernetes-api-server", "", "Kubernetes API server URL")
	controllerCmd.Flags().StringVar(&kubernetesCAFilePath, "kubernetes-ca-file-path", "", "Kubernetes CA file path")
	controllerCmd.Flags().StringVar(&kubernetesCertFilePath, "kubernetes-cert-file-path", "", "Kubernetes cert file path")
	controllerCmd.Flags().StringVar(&kubernetesKeyFilePath, "kubernetes-key-file-path", "", "Kubernetes key file path")

	controllerCmd.Flags().StringVar(&listenAddress, "listen-address", "127.0.0.1:8000", "Listen address for server")
}

var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Start the controller",
	Run:   controllerRun,
}

func controllerRun(cmd *cobra.Command, args []string) {
	config := controller.Config{
		KubernetesAPIServer:    kubernetesAPIServer,
		KubernetesCAFilePath:   kubernetesCAFilePath,
		KubernetesCertFilePath: kubernetesCertFilePath,
		KubernetesKeyFilePath:  kubernetesKeyFilePath,

		ListenAddress: listenAddress,
	}

	controller := controller.New(config)
	controller.Start()
}
