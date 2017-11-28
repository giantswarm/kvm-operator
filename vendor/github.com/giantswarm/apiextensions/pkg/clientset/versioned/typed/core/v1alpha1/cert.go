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

// CertsGetter has a method to return a CertInterface.
// A group's client should implement this interface.
type CertsGetter interface {
	Certs(namespace string) CertInterface
}

// CertInterface has methods to work with Cert resources.
type CertInterface interface {
	Create(*v1alpha1.Cert) (*v1alpha1.Cert, error)
	Update(*v1alpha1.Cert) (*v1alpha1.Cert, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Cert, error)
	List(opts v1.ListOptions) (*v1alpha1.CertList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Cert, err error)
	CertExpansion
}

// certs implements CertInterface
type certs struct {
	client rest.Interface
	ns     string
}

// newCerts returns a Certs
func newCerts(c *CoreV1alpha1Client, namespace string) *certs {
	return &certs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the cert, and returns the corresponding cert object, and an error if there is any.
func (c *certs) Get(name string, options v1.GetOptions) (result *v1alpha1.Cert, err error) {
	result = &v1alpha1.Cert{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("certs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Certs that match those selectors.
func (c *certs) List(opts v1.ListOptions) (result *v1alpha1.CertList, err error) {
	result = &v1alpha1.CertList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("certs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested certs.
func (c *certs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("certs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a cert and creates it.  Returns the server's representation of the cert, and an error, if there is any.
func (c *certs) Create(cert *v1alpha1.Cert) (result *v1alpha1.Cert, err error) {
	result = &v1alpha1.Cert{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("certs").
		Body(cert).
		Do().
		Into(result)
	return
}

// Update takes the representation of a cert and updates it. Returns the server's representation of the cert, and an error, if there is any.
func (c *certs) Update(cert *v1alpha1.Cert) (result *v1alpha1.Cert, err error) {
	result = &v1alpha1.Cert{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("certs").
		Name(cert.Name).
		Body(cert).
		Do().
		Into(result)
	return
}

// Delete takes name of the cert and deletes it. Returns an error if one occurs.
func (c *certs) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("certs").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *certs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("certs").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched cert.
func (c *certs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Cert, err error) {
	result = &v1alpha1.Cert{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("certs").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
