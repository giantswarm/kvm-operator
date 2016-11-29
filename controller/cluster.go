package controller

import 	(
	"bytes"
	"errors"
	"text/template"
)

type ClusterObj interface {
	Create() error
	Delete() error
}

type clusterObj struct {
	master *master
	configMap *configMap
	worker *worker
	flannelClient *flannelClient
	namespace *namespace
}


func NewCluster(config ClusterConfig) (ClusterObj, error) {
	if config.ClusterID == "" {
		return nil, errors.New("cluster ID must not be empty")
	}
	if config.KubernetesClient == nil {
		return nil, errors.New("kubernetes client must not be empty")
	}
	if config.Namespace == "" {
		return nil, errors.New("namespace must not be empty")
	}
	if config.Replicas == int32(0) {
		return nil, errors.New("replicas must not be empty")
	}

	newClusterObj := &clusterObj{
		configMap: &configMap{
								ClusterConfig: config,
		},
		master: &master{
								ClusterConfig: config,
							},
		worker: &worker{
								ClusterConfig: config,
							},
		flannelClient: &flannelClient{
								ClusterConfig: config,
							},
		namespace: &namespace{
								ClusterConfig: config,
							},
	}

	return newClusterObj, nil
}

// New creates a new configured worker object.
func (c *clusterObj) Create() error {
	if err := c.configMap.Create(); err != nil {
		return err
	}

	if err := c.namespace.Create(); err != nil {
		return err
	}

	if err := c.master.Create(); err != nil {
		return err
	}

	if err := c.worker.Create(); err != nil {
		return err
	}

	return nil
}

func (c *clusterObj) Delete() error {
	if err := c.namespace.Delete(); err != nil {
		return err
	}

	return nil
}

func GetNamespaceNameForCluster(config ClusterConfig) string {
	newNamespace := &namespace{
		ClusterConfig: config,
	}

	return newNamespace.GetNamespace()
}

func ExecTemplate(t string, obj interface{}) (string, error) {
	var result bytes.Buffer

	tmpl, err := template.New("component").Parse(t)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(&result, obj)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}
