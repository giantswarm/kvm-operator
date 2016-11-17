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
	KubernetesAPIServer    string
	KubernetesCAFilePath   string
	KubernetesCertFilePath string
	KubernetesKeyFilePath  string

	ListenAddress string
}

func New(config Config) Controller {
	clientset := kubernetes.NewForConfigOrDie(
		&rest.Config{
			Host: config.KubernetesAPIServer,
			TLSClientConfig: rest.TLSClientConfig{
				CAFile:   config.KubernetesCAFilePath,
				CertFile: config.KubernetesCertFilePath,
				KeyFile:  config.KubernetesKeyFilePath,
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
	if err := c.createClusterResource(); err != nil {
		log.Fatalln("Could not create cluster resource:", err)
	}

	c.startServer()
}
