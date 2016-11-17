package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(controllerCmd)
}

var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Start the controller",
	Run:   controllerRun,
}

func controllerRun(cmd *cobra.Command, args []string) {
	log.Println("Starting controller")
}
