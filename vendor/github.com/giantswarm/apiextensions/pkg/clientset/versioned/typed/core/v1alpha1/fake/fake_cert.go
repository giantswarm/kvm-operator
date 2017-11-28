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

package fake

import (
	v1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeCerts implements CertInterface
type FakeCerts struct {
	Fake *FakeCoreV1alpha1
	ns   string
}

var certsResource = schema.GroupVersionResource{Group: "core.giantswarm.io", Version: "v1alpha1", Resource: "certs"}

var certsKind = schema.GroupVersionKind{Group: "core.giantswarm.io", Version: "v1alpha1", Kind: "Cert"}

// Get takes name of the cert, and returns the corresponding cert object, and an error if there is any.
func (c *FakeCerts) Get(name string, options v1.GetOptions) (result *v1alpha1.Cert, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(certsResource, c.ns, name), &v1alpha1.Cert{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Cert), err
}

// List takes label and field selectors, and returns the list of Certs that match those selectors.
func (c *FakeCerts) List(opts v1.ListOptions) (result *v1alpha1.CertList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(certsResource, certsKind, c.ns, opts), &v1alpha1.CertList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.CertList{}
	for _, item := range obj.(*v1alpha1.CertList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested certs.
func (c *FakeCerts) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(certsResource, c.ns, opts))

}

// Create takes the representation of a cert and creates it.  Returns the server's representation of the cert, and an error, if there is any.
func (c *FakeCerts) Create(cert *v1alpha1.Cert) (result *v1alpha1.Cert, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(certsResource, c.ns, cert), &v1alpha1.Cert{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Cert), err
}

// Update takes the representation of a cert and updates it. Returns the server's representation of the cert, and an error, if there is any.
func (c *FakeCerts) Update(cert *v1alpha1.Cert) (result *v1alpha1.Cert, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(certsResource, c.ns, cert), &v1alpha1.Cert{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Cert), err
}

// Delete takes name of the cert and deletes it. Returns an error if one occurs.
func (c *FakeCerts) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(certsResource, c.ns, name), &v1alpha1.Cert{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCerts) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(certsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.CertList{})
	return err
}

// Patch applies the patch and returns the patched cert.
func (c *FakeCerts) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Cert, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(certsResource, c.ns, name, data, subresources...), &v1alpha1.Cert{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Cert), err
}
