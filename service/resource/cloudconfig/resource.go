package cloudconfig

import (
	"fmt"
	"reflect"

	"github.com/giantswarm/certificatetpr"
	cloudconfig "github.com/giantswarm/k8scloudconfig"
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
	FileOwner      = "root:root"
	FilePermission = 0700
	KeyUserData    = "user_data"
	// Name is the identifier of the resource.
	Name         = "cloudconfig"
	PrefixMaster = "master"
	PrefixWorker = "worker"
)

// Config represents the configuration used to create a new cloud config resource.
type Config struct {
	// Dependencies.
	CertWatcher *certificatetpr.Service
	K8sClient   kubernetes.Interface
	Logger      micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new cloud config
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		CertWatcher: nil,
		K8sClient:   nil,
		Logger:      nil,
	}
}

// Resource implements the cloud config resource.
type Resource struct {
	// Dependencies.
	certWatcher *certificatetpr.Service
	k8sClient   kubernetes.Interface
	logger      micrologger.Logger
}

// New creates a new configured cloud config resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.CertWatcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CertWatcher must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newService := &Resource{
		// Dependencies.
		certWatcher: config.CertWatcher,
		k8sClient:   config.K8sClient,
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

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "looking for config maps in the Kubernetes API", "operation", "GetCurrentState")

	// Lookup the current state of the configmaps.
	var currentConfigMaps []*apiv1.ConfigMap

	namespace := key.ClusterNamespace(customObject)
	configMapNames := getConfigMapNames(customObject)

	for _, name := range configMapNames {
		configMap, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Get(name, apismetav1.GetOptions{})
		if err != nil {
			return nil, microerror.Mask(err)
		}
		currentConfigMaps = append(currentConfigMaps, configMap)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d config maps in the Kubernetes API", len(currentConfigMaps)))

	return currentConfigMaps, nil
}

func (r *Resource) GetDesiredState(obj interface{}) (interface{}, error) {
	customObject, err := interfaceToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "making up new config maps", "operation", "GetDesiredState")

	// Lookup the desired state of the config maps to have a reference of data how
	// it should be.
	desiredConfigMaps, err := r.newConfigMaps(customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("made up %d config maps", len(desiredConfigMaps)), "operation", "GetDesiredState")

	return desiredConfigMaps, nil
}

func (r *Resource) GetCreateState(obj, cur, des interface{}) (interface{}, error) {
	customObject, err := interfaceToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentConfigMaps, err := interfaceToConfigMaps(cur)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredConfigMaps, err := interfaceToConfigMaps(des)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which config maps to create", "operation", "GetCreateState")

	// Find anything which is in the desired state but not in the current state.
	// This lets us drive the current state towards the desired state, because
	// everything we find here is supposed to be created. In case a config map is
	// in the current and the desired state we check if their data is equal. If
	// the data differs the config map is supposed to be updated to bring the
	// current state into the desired state.
	configMapsToCreate := currentConfigMaps
	for _, desiredConfigMap := range desiredConfigMaps {
		if !containsConfigMap(currentConfigMaps, desiredConfigMap) {
			configMapsToCreate = append(configMapsToCreate, desiredConfigMap)
		} else {
			currentConfigMap, err := getConfigMapByName(currentConfigMaps, desiredConfigMap.Name)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			if !reflect.DeepEqual(desiredConfigMap.Data, currentConfigMap.Data) {
				configMapsToCreate = append(configMapsToCreate, desiredConfigMap)
			}
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d config maps to create", len(configMapsToCreate)), "operation", "GetCreateState")

	return configMapsToCreate, nil
}

func (r *Resource) GetDeleteState(obj, cur, des interface{}) (interface{}, error) {
	customObject, err := interfaceToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentConfigMaps, err := interfaceToConfigMaps(cur)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredConfigMaps, err := interfaceToConfigMaps(des)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which config maps to delete", "operation", "GetDeleteState")

	// Find anything which is in the current state but not in the desired state.
	// This lets us drive the current state towards the desired state, because
	// everything we find here is supposed to be deleted.
	var configMapsToDelete []*apiv1.ConfigMap
	for _, currentConfigMap := range currentConfigMaps {
		if !containsConfigMap(desiredConfigMaps, currentConfigMap) {
			configMapsToDelete = append(configMapsToDelete, currentConfigMap)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d config maps to delete", len(configMapsToDelete)), "operation", "GetDeleteState")

	return configMapsToDelete, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ProcessCreateState(obj, cre interface{}) error {
	customObject, err := interfaceToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	configMapsToCreate, err := interfaceToConfigMaps(cre)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "creating config maps in the Kubernetes API", "operation", "ProcessCreateState")

	// Create the config maps in the Kubernetes API.
	namespace := key.ClusterNamespace(customObject)
	for _, configMap := range configMapsToCreate {
		_, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Create(configMap)
		if apierrors.IsAlreadyExists(err) {
			_, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Update(configMap)
			if err != nil {
				return microerror.Mask(err)
			}
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "created config maps in the Kubernetes API", "operation", "ProcessCreateState")

	return nil
}

func (r *Resource) ProcessDeleteState(obj, del interface{}) error {
	customObject, err := interfaceToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	configMapsToDelete, err := interfaceToConfigMaps(del)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "deleting config maps in the Kubernetes API", "operation", "ProcessDeleteState")

	// Create the config maps in the Kubernetes API.
	namespace := key.ClusterNamespace(customObject)
	for _, configMap := range configMapsToDelete {
		err := r.k8sClient.CoreV1().ConfigMaps(namespace).Delete(configMap.Name, &apismetav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "deleted config maps in the Kubernetes API", "operation", "ProcessDeleteState")

	return nil
}

func (r *Resource) Underlying() framework.Resource {
	return r
}

// newConfigMap creates a new Kubernetes configmap using the provided
// information. customObject is used for name and label creation. params serves
// as structure being injected into the template execution to interpolate
// variables. prefix can be either "master" or "worker" and is used to prefix
// the configmap name.
func (r *Resource) newConfigMap(customObject kvmtpr.CustomObject, template string, params cloudconfig.Params, prefix string) (*apiv1.ConfigMap, error) {
	var err error

	var newCloudConfig *cloudconfig.CloudConfig
	{
		newCloudConfig, err = cloudconfig.NewCloudConfig(template, params)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		err = newCloudConfig.ExecuteTemplate()
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var newConfigMap *apiv1.ConfigMap
	{
		newConfigMap = &apiv1.ConfigMap{
			ObjectMeta: apismetav1.ObjectMeta{
				Name: key.ConfigMapName(customObject, params.Node, prefix),
				Labels: map[string]string{
					"cluster":  key.ClusterID(customObject),
					"customer": key.ClusterCustomer(customObject),
				},
			},
			Data: map[string]string{
				KeyUserData: newCloudConfig.Base64(),
			},
		}
	}

	return newConfigMap, nil
}

func (r *Resource) newConfigMaps(customObject kvmtpr.CustomObject) ([]*apiv1.ConfigMap, error) {
	var configMaps []*apiv1.ConfigMap

	certs, err := r.certWatcher.SearchCerts(customObject.Spec.Cluster.Cluster.ID)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, node := range customObject.Spec.Cluster.Masters {
		newExtension := &MasterExtension{
			certs: certs,
		}

		var params cloudconfig.Params
		{
			params.Cluster = customObject.Spec.Cluster
			params.Extension = newExtension
			params.Node = node
		}

		configMap, err := r.newConfigMap(customObject, cloudconfig.MasterTemplate, params, PrefixMaster)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMaps = append(configMaps, configMap)
	}

	for _, node := range customObject.Spec.Cluster.Workers {
		newExtension := &WorkerExtension{
			certs: certs,
		}

		var params cloudconfig.Params
		{
			params.Cluster = customObject.Spec.Cluster
			params.Extension = newExtension
			params.Node = node
		}

		configMap, err := r.newConfigMap(customObject, cloudconfig.WorkerTemplate, params, PrefixWorker)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMaps = append(configMaps, configMap)
	}

	return configMaps, nil
}

func containsConfigMap(list []*apiv1.ConfigMap, item *apiv1.ConfigMap) bool {
	for _, l := range list {
		if l.Name == item.Name {
			return true
		}
	}

	return false
}

func getConfigMapByName(list []*apiv1.ConfigMap, name string) (*apiv1.ConfigMap, error) {
	for _, l := range list {
		if l.Name == name {
			return l, nil
		}
	}

	return nil, microerror.Maskf(notFoundError, "could not find config map '%s'", name)
}

func getConfigMapNames(customObject kvmtpr.CustomObject) []string {
	var names []string

	for _, node := range customObject.Spec.Cluster.Masters {
		name := key.ConfigMapName(customObject, node, PrefixMaster)
		names = append(names, name)
	}

	for _, node := range customObject.Spec.Cluster.Workers {
		name := key.ConfigMapName(customObject, node, PrefixWorker)
		names = append(names, name)
	}

	return names
}

func interfaceToCustomObject(v interface{}) (kvmtpr.CustomObject, error) {
	customObjectPointer, ok := v.(*kvmtpr.CustomObject)
	if !ok {
		return kvmtpr.CustomObject{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, v)
	}
	customObject := *customObjectPointer

	return customObject, nil
}

func interfaceToConfigMaps(v interface{}) ([]*apiv1.ConfigMap, error) {
	configMaps, ok := v.([]*apiv1.ConfigMap)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*apiv1.ConfigMap{}, v)
	}

	return configMaps, nil
}
