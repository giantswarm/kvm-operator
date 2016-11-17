package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "cluster-controller",
	Short: "cluster-controller handles Kubernetes clusters running on a Kubernetes cluster",
}
