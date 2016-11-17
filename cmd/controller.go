package cmd

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/cluster-controller/controller"
)

var listenAddress string

func init() {
	RootCmd.AddCommand(controllerCmd)

	controllerCmd.Flags().StringVar(&listenAddress, "listen-address", "127.0.0.1:8000", "Listen address for server")
}

var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Start the controller",
	Run:   controllerRun,
}

func controllerRun(cmd *cobra.Command, args []string) {
	controller := controller.New(listenAddress)
	controller.Start()
}
