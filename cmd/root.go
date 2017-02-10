package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "kvm-operator",
	Short: "kvm-operator handles Kubernetes clusters running on a Kubernetes cluster",
}
