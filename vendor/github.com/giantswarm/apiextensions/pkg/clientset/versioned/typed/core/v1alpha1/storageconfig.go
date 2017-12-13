/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	v1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	scheme "github.com/giantswarm/apiextensions/pkg/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// StorageConfigsGetter has a method to return a StorageConfigInterface.
// A group's client should implement this interface.
type StorageConfigsGetter interface {
	StorageConfigs(namespace string) StorageConfigInterface
}

// StorageConfigInterface has methods to work with StorageConfig resources.
type StorageConfigInterface interface {
	Create(*v1alpha1.StorageConfig) (*v1alpha1.StorageConfig, error)
	Update(*v1alpha1.StorageConfig) (*v1alpha1.StorageConfig, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.StorageConfig, error)
	List(opts v1.ListOptions) (*v1alpha1.StorageConfigList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.StorageConfig, err error)
	StorageConfigExpansion
}

// storageConfigs implements StorageConfigInterface
type storageConfigs struct {
	client rest.Interface
	ns     string
}

// newStorageConfigs returns a StorageConfigs
func newStorageConfigs(c *CoreV1alpha1Client, namespace string) *storageConfigs {
	return &storageConfigs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the storageConfig, and returns the corresponding storageConfig object, and an error if there is any.
func (c *storageConfigs) Get(name string, options v1.GetOptions) (result *v1alpha1.StorageConfig, err error) {
	result = &v1alpha1.StorageConfig{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("storageconfigs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of StorageConfigs that match those selectors.
func (c *storageConfigs) List(opts v1.ListOptions) (result *v1alpha1.StorageConfigList, err error) {
	result = &v1alpha1.StorageConfigList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("storageconfigs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested storageConfigs.
func (c *storageConfigs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("storageconfigs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a storageConfig and creates it.  Returns the server's representation of the storageConfig, and an error, if there is any.
func (c *storageConfigs) Create(storageConfig *v1alpha1.StorageConfig) (result *v1alpha1.StorageConfig, err error) {
	result = &v1alpha1.StorageConfig{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("storageconfigs").
		Body(storageConfig).
		Do().
		Into(result)
	return
}

// Update takes the representation of a storageConfig and updates it. Returns the server's representation of the storageConfig, and an error, if there is any.
func (c *storageConfigs) Update(storageConfig *v1alpha1.StorageConfig) (result *v1alpha1.StorageConfig, err error) {
	result = &v1alpha1.StorageConfig{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("storageconfigs").
		Name(storageConfig.Name).
		Body(storageConfig).
		Do().
		Into(result)
	return
}

// Delete takes name of the storageConfig and deletes it. Returns an error if one occurs.
func (c *storageConfigs) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("storageconfigs").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *storageConfigs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("storageconfigs").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched storageConfig.
func (c *storageConfigs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.StorageConfig, err error) {
	result = &v1alpha1.StorageConfig{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("storageconfigs").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
