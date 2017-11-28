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
	v1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	scheme "github.com/giantswarm/apiextensions/pkg/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// AzuresGetter has a method to return a AzureInterface.
// A group's client should implement this interface.
type AzuresGetter interface {
	Azures(namespace string) AzureInterface
}

// AzureInterface has methods to work with Azure resources.
type AzureInterface interface {
	Create(*v1alpha1.Azure) (*v1alpha1.Azure, error)
	Update(*v1alpha1.Azure) (*v1alpha1.Azure, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Azure, error)
	List(opts v1.ListOptions) (*v1alpha1.AzureList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Azure, err error)
	AzureExpansion
}

// azures implements AzureInterface
type azures struct {
	client rest.Interface
	ns     string
}

// newAzures returns a Azures
func newAzures(c *ClusterV1alpha1Client, namespace string) *azures {
	return &azures{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the azure, and returns the corresponding azure object, and an error if there is any.
func (c *azures) Get(name string, options v1.GetOptions) (result *v1alpha1.Azure, err error) {
	result = &v1alpha1.Azure{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("azures").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Azures that match those selectors.
func (c *azures) List(opts v1.ListOptions) (result *v1alpha1.AzureList, err error) {
	result = &v1alpha1.AzureList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("azures").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested azures.
func (c *azures) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("azures").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a azure and creates it.  Returns the server's representation of the azure, and an error, if there is any.
func (c *azures) Create(azure *v1alpha1.Azure) (result *v1alpha1.Azure, err error) {
	result = &v1alpha1.Azure{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("azures").
		Body(azure).
		Do().
		Into(result)
	return
}

// Update takes the representation of a azure and updates it. Returns the server's representation of the azure, and an error, if there is any.
func (c *azures) Update(azure *v1alpha1.Azure) (result *v1alpha1.Azure, err error) {
	result = &v1alpha1.Azure{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("azures").
		Name(azure.Name).
		Body(azure).
		Do().
		Into(result)
	return
}

// Delete takes name of the azure and deletes it. Returns an error if one occurs.
func (c *azures) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("azures").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *azures) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("azures").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched azure.
func (c *azures) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Azure, err error) {
	result = &v1alpha1.Azure{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("azures").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
