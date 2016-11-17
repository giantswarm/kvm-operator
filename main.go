package main

import (
	"log"

	"github.com/giantswarm/cluster-controller/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
