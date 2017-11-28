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
	v1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeKVMs implements KVMInterface
type FakeKVMs struct {
	Fake *FakeClusterV1alpha1
	ns   string
}

var kvmsResource = schema.GroupVersionResource{Group: "cluster.giantswarm.io", Version: "v1alpha1", Resource: "kvms"}

var kvmsKind = schema.GroupVersionKind{Group: "cluster.giantswarm.io", Version: "v1alpha1", Kind: "KVM"}

// Get takes name of the kVM, and returns the corresponding kVM object, and an error if there is any.
func (c *FakeKVMs) Get(name string, options v1.GetOptions) (result *v1alpha1.KVM, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(kvmsResource, c.ns, name), &v1alpha1.KVM{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KVM), err
}

// List takes label and field selectors, and returns the list of KVMs that match those selectors.
func (c *FakeKVMs) List(opts v1.ListOptions) (result *v1alpha1.KVMList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(kvmsResource, kvmsKind, c.ns, opts), &v1alpha1.KVMList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.KVMList{}
	for _, item := range obj.(*v1alpha1.KVMList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested kVMs.
func (c *FakeKVMs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(kvmsResource, c.ns, opts))

}

// Create takes the representation of a kVM and creates it.  Returns the server's representation of the kVM, and an error, if there is any.
func (c *FakeKVMs) Create(kVM *v1alpha1.KVM) (result *v1alpha1.KVM, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(kvmsResource, c.ns, kVM), &v1alpha1.KVM{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KVM), err
}

// Update takes the representation of a kVM and updates it. Returns the server's representation of the kVM, and an error, if there is any.
func (c *FakeKVMs) Update(kVM *v1alpha1.KVM) (result *v1alpha1.KVM, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(kvmsResource, c.ns, kVM), &v1alpha1.KVM{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KVM), err
}

// Delete takes name of the kVM and deletes it. Returns an error if one occurs.
func (c *FakeKVMs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(kvmsResource, c.ns, name), &v1alpha1.KVM{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeKVMs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(kvmsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.KVMList{})
	return err
}

// Patch applies the patch and returns the patched kVM.
func (c *FakeKVMs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.KVM, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(kvmsResource, c.ns, name, data, subresources...), &v1alpha1.KVM{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KVM), err
}
