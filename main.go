package main

import (
	"log"

	"github.com/giantswarm/kvm-operator/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
