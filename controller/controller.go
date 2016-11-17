package controller

import (
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	ClusterThirdPartyResourceName = "cluster.giantswarm.io"
)

type Controller interface {
	Start()
}

type controller struct {
	clientset *kubernetes.Clientset
	config    Config
}

type Config struct {
	APIServer    string
	CAFilePath   string
	CertFilePath string
	KeyFilePath  string

	ListenAddress string
}

func New(config Config) Controller {
	clientset := kubernetes.NewForConfigOrDie(
		&rest.Config{
			Host: config.APIServer,
			TLSClientConfig: rest.TLSClientConfig{
				CAFile:   config.CAFilePath,
				CertFile: config.CertFilePath,
				KeyFile:  config.KeyFilePath,
			},
		},
	)

	return &controller{
		clientset: clientset,
		config:    config,
	}
}

// Start starts the server.
func (c *controller) Start() {
	go c.startServer()

	if err := c.createClusterResource(); err != nil {
		log.Fatalln("could not create cluster resource:", err)
	}

	// dummy - the controller loops will replace this
	for {
		select {}
	}
}
