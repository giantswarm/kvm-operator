package namespace

import (
	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/service/key"
)

const (
	// Name is the identifier of the resource.
	Name = "namespace"
)

// Config represents the configuration used to create a new cloud config resource.
type Config struct {
	// Dependencies.
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new cloud config
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		K8sClient: nil,
		Logger:    nil,
	}
}

// Resource implements the cloud config resource.
type Resource struct {
	// Dependencies.
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

// New creates a new configured cloud config resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newService := &Resource{
		// Dependencies.
		k8sClient: config.K8sClient,
		logger: config.Logger.With(
			"resource", Name,
		),
	}

	return newService, nil
}

func (r *Resource) GetCurrentState(obj interface{}) (interface{}, error) {
	customObject, err := interfaceToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "looking for the namespace in the Kubernetes API")

	// Lookup the current state of the namespace.
	namespaceName := key.ClusterNamespace(customObject)
	namespace, err := r.k8sClient.CoreV1().Namespaces().Get(namespaceName, apismetav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		// fall through
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	if namespace == nil {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "did not found the namespace in the Kubernetes API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "found the namespace in the Kubernetes API")
	}

	return namespace, nil
}

func (r *Resource) GetDesiredState(obj interface{}) (interface{}, error) {
	customObject, err := interfaceToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "computing the new namespace")

	// Make up the desired state of the namespace to have a reference of data how
	// it should be.
	namespace := &apiv1.Namespace{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name: key.ClusterNamespace(customObject),
			Labels: map[string]string{
				"cluster":  key.ClusterID(customObject),
				"customer": key.ClusterCustomer(customObject),
			},
		},
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "computed the new namespace")

	return namespace, nil
}

func (r *Resource) GetCreateState(obj, cur, des interface{}) (interface{}, error) {
	customObject, err := interfaceToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentNamespace, err := interfaceToNamespace(cur)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredNamespace, err := interfaceToNamespace(des)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out if the namespace has to be created")

	var namespaceToCreate *apiv1.Namespace
	if currentNamespace == nil {
		namespaceToCreate = desiredNamespace
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "found out if the namespace has to be created")

	return namespaceToCreate, nil
}

func (r *Resource) GetDeleteState(obj, cur, des interface{}) (interface{}, error) {
	customObject, err := interfaceToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentNamespace, err := interfaceToNamespace(cur)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredNamespace, err := interfaceToNamespace(des)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out if the namespace has to be deleted")

	var namespaceToDelete *apiv1.Namespace
	if currentNamespace != nil {
		namespaceToDelete = desiredNamespace
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "found out if the namespace has to be deleted")

	return namespaceToDelete, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ProcessCreateState(obj, cre interface{}) error {
	customObject, err := interfaceToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	namespaceToCreate, err := interfaceToNamespace(cre)
	if err != nil {
		return microerror.Mask(err)
	}

	if namespaceToCreate != nil {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "creating the namespace in the Kubernetes API")

		_, err = r.k8sClient.CoreV1().Namespaces().Create(namespaceToCreate)
		if apierrors.IsAlreadyExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "created the namespace in the Kubernetes API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "the namespace does already exist in the Kubernetes API")
	}

	return nil
}

func (r *Resource) ProcessDeleteState(obj, del interface{}) error {
	customObject, err := interfaceToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	namespaceToDelete, err := interfaceToNamespace(del)
	if err != nil {
		return microerror.Mask(err)
	}

	if namespaceToDelete != nil {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "deleting the namespace in the Kubernetes API")

		err = r.k8sClient.CoreV1().Namespaces().Delete(namespaceToDelete.Name, &apismetav1.DeleteOptions{})
		if apierrors.IsAlreadyExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "deleted the namespace in the Kubernetes API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "the namespace does not exist in the Kubernetes API")
	}

	return nil
}

func (r *Resource) Underlying() framework.Resource {
	return r
}

func interfaceToCustomObject(v interface{}) (kvmtpr.CustomObject, error) {
	customObjectPointer, ok := v.(*kvmtpr.CustomObject)
	if !ok {
		return kvmtpr.CustomObject{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, v)
	}
	customObject := *customObjectPointer

	return customObject, nil
}

func interfaceToNamespace(v interface{}) (*apiv1.Namespace, error) {
	if v == nil {
		return nil, nil
	}

	namespace, ok := v.(*apiv1.Namespace)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &apiv1.Namespace{}, v)
	}

	return namespace, nil
}
