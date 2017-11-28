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

// KVMsGetter has a method to return a KVMInterface.
// A group's client should implement this interface.
type KVMsGetter interface {
	KVMs(namespace string) KVMInterface
}

// KVMInterface has methods to work with KVM resources.
type KVMInterface interface {
	Create(*v1alpha1.KVM) (*v1alpha1.KVM, error)
	Update(*v1alpha1.KVM) (*v1alpha1.KVM, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.KVM, error)
	List(opts v1.ListOptions) (*v1alpha1.KVMList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.KVM, err error)
	KVMExpansion
}

// kVMs implements KVMInterface
type kVMs struct {
	client rest.Interface
	ns     string
}

// newKVMs returns a KVMs
func newKVMs(c *ClusterV1alpha1Client, namespace string) *kVMs {
	return &kVMs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the kVM, and returns the corresponding kVM object, and an error if there is any.
func (c *kVMs) Get(name string, options v1.GetOptions) (result *v1alpha1.KVM, err error) {
	result = &v1alpha1.KVM{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("kvms").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of KVMs that match those selectors.
func (c *kVMs) List(opts v1.ListOptions) (result *v1alpha1.KVMList, err error) {
	result = &v1alpha1.KVMList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("kvms").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested kVMs.
func (c *kVMs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("kvms").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a kVM and creates it.  Returns the server's representation of the kVM, and an error, if there is any.
func (c *kVMs) Create(kVM *v1alpha1.KVM) (result *v1alpha1.KVM, err error) {
	result = &v1alpha1.KVM{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("kvms").
		Body(kVM).
		Do().
		Into(result)
	return
}

// Update takes the representation of a kVM and updates it. Returns the server's representation of the kVM, and an error, if there is any.
func (c *kVMs) Update(kVM *v1alpha1.KVM) (result *v1alpha1.KVM, err error) {
	result = &v1alpha1.KVM{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("kvms").
		Name(kVM.Name).
		Body(kVM).
		Do().
		Into(result)
	return
}

// Delete takes name of the kVM and deletes it. Returns an error if one occurs.
func (c *kVMs) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("kvms").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *kVMs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("kvms").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched kVM.
func (c *kVMs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.KVM, err error) {
	result = &v1alpha1.KVM{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("kvms").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
