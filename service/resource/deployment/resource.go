package deployment

import (
	"fmt"

	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/giantswarm/kvm-operator/service/key"
)

const (
	MasterID = "master"
	// Name is the identifier of the resource.
	Name     = "deployment"
	WorkerID = "worker"
)

// Config represents the configuration used to create a new deployment resource.
type Config struct {
	// Dependencies.
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new deployment
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		K8sClient: nil,
		Logger:    nil,
	}
}

// Resource implements the deployment resource.
type Resource struct {
	// Dependencies.
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

// New creates a new configured deployment resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newResource := &Resource{
		// Dependencies.
		k8sClient: config.K8sClient,
		logger: config.Logger.With(
			"resource", Name,
		),
	}

	return newResource, nil
}

func (r *Resource) GetCurrentState(obj interface{}) (interface{}, error) {
	customObject, err := toCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "looking for deployments in the Kubernetes API")

	var deployments []*v1beta1.Deployment

	namespace := key.ClusterNamespace(customObject)
	deploymentNames := getDeploymentNames(customObject)

	for _, name := range deploymentNames {
		manifest, err := r.k8sClient.Extensions().Deployments(namespace).Get(name, apismetav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.Log("cluster", key.ClusterID(customObject), "debug", "did not find a deployment in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Log("cluster", key.ClusterID(customObject), "debug", "found a deployment in the Kubernetes API")
			deployments = append(deployments, manifest)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d deployments in the Kubernetes API", len(deployments)))

	return deployments, nil
}

func (r *Resource) GetDesiredState(obj interface{}) (interface{}, error) {
	customObject, err := toCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "computing the new deployments")

	var deployments []*v1beta1.Deployment

	{
		masterDeployments, err := newMasterDeployments(customObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		deployments = append(deployments, masterDeployments...)

		workerDeployments, err := newWorkerDeployments(customObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		deployments = append(deployments, workerDeployments...)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("computed the %d new deployments", len(deployments)))

	return deployments, nil
}

func (r *Resource) GetCreateState(obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := toCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentDeployments, err := toDeployments(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredDeployments, err := toDeployments(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which deployments have to be created")

	var deploymentsToCreate []*v1beta1.Deployment

	for _, desiredDeployment := range desiredDeployments {
		if !containsDeployment(currentDeployments, desiredDeployment) {
			deploymentsToCreate = append(deploymentsToCreate, desiredDeployment)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d deployments that have to be created", len(deploymentsToCreate)))

	return deploymentsToCreate, nil
}

func (r *Resource) GetDeleteState(obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := toCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentDeployments, err := toDeployments(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredDeployments, err := toDeployments(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which deployments have to be deleted")

	var deploymentsToDelete []*v1beta1.Deployment

	for _, currentDeployment := range currentDeployments {
		if containsDeployment(desiredDeployments, currentDeployment) {
			deploymentsToDelete = append(deploymentsToDelete, currentDeployment)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d deployments that have to be deleted", len(deploymentsToDelete)))

	return deploymentsToDelete, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ProcessCreateState(obj, createState interface{}) error {
	customObject, err := toCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	deploymentsToCreate, err := toDeployments(createState)
	if err != nil {
		return microerror.Mask(err)
	}

	if deploymentsToCreate != nil {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "creating the deployments in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, deployment := range deploymentsToCreate {
			_, err := r.k8sClient.Extensions().Deployments(namespace).Create(deployment)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "created the deployments in the Kubernetes API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "the deployments do already exist in the Kubernetes API")
	}

	return nil
}

func (r *Resource) ProcessDeleteState(obj, deleteState interface{}) error {
	customObject, err := toCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	deploymentsToDelete, err := toDeployments(deleteState)
	if err != nil {
		return microerror.Mask(err)
	}

	if deploymentsToDelete != nil {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "deleting the deployments in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, deployment := range deploymentsToDelete {
			err := r.k8sClient.Extensions().Deployments(namespace).Delete(deployment.Name, &apismetav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "deleted the deployments in the Kubernetes API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "the deployments do not exist in the Kubernetes API")
	}

	return nil
}

func (r *Resource) Underlying() framework.Resource {
	return r
}

func containsDeployment(list []*v1beta1.Deployment, item *v1beta1.Deployment) bool {
	for _, l := range list {
		if l.Name == item.Name {
			return true
		}
	}

	return false
}

func getDeploymentNames(customObject kvmtpr.CustomObject) []string {
	var names []string

	for _, node := range customObject.Spec.Cluster.Masters {
		names = append(names, "master-"+node.ID)
	}

	for _, node := range customObject.Spec.Cluster.Workers {
		names = append(names, "worker-"+node.ID)
	}

	return names
}

func toCustomObject(v interface{}) (kvmtpr.CustomObject, error) {
	customObjectPointer, ok := v.(*kvmtpr.CustomObject)
	if !ok {
		return kvmtpr.CustomObject{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, v)
	}
	customObject := *customObjectPointer

	return customObject, nil
}

func toDeployments(v interface{}) ([]*v1beta1.Deployment, error) {
	if v == nil {
		return nil, nil
	}

	deployments, ok := v.([]*v1beta1.Deployment)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*v1beta1.Deployment{}, v)
	}

	return deployments, nil
}
