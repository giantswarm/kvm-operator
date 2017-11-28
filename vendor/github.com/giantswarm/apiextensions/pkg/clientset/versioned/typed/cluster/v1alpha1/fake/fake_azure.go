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

// FakeAzures implements AzureInterface
type FakeAzures struct {
	Fake *FakeClusterV1alpha1
	ns   string
}

var azuresResource = schema.GroupVersionResource{Group: "cluster.giantswarm.io", Version: "v1alpha1", Resource: "azures"}

var azuresKind = schema.GroupVersionKind{Group: "cluster.giantswarm.io", Version: "v1alpha1", Kind: "Azure"}

// Get takes name of the azure, and returns the corresponding azure object, and an error if there is any.
func (c *FakeAzures) Get(name string, options v1.GetOptions) (result *v1alpha1.Azure, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(azuresResource, c.ns, name), &v1alpha1.Azure{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Azure), err
}

// List takes label and field selectors, and returns the list of Azures that match those selectors.
func (c *FakeAzures) List(opts v1.ListOptions) (result *v1alpha1.AzureList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(azuresResource, azuresKind, c.ns, opts), &v1alpha1.AzureList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.AzureList{}
	for _, item := range obj.(*v1alpha1.AzureList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested azures.
func (c *FakeAzures) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(azuresResource, c.ns, opts))

}

// Create takes the representation of a azure and creates it.  Returns the server's representation of the azure, and an error, if there is any.
func (c *FakeAzures) Create(azure *v1alpha1.Azure) (result *v1alpha1.Azure, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(azuresResource, c.ns, azure), &v1alpha1.Azure{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Azure), err
}

// Update takes the representation of a azure and updates it. Returns the server's representation of the azure, and an error, if there is any.
func (c *FakeAzures) Update(azure *v1alpha1.Azure) (result *v1alpha1.Azure, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(azuresResource, c.ns, azure), &v1alpha1.Azure{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Azure), err
}

// Delete takes name of the azure and deletes it. Returns an error if one occurs.
func (c *FakeAzures) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(azuresResource, c.ns, name), &v1alpha1.Azure{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeAzures) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(azuresResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.AzureList{})
	return err
}

// Patch applies the patch and returns the patched azure.
func (c *FakeAzures) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Azure, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(azuresResource, c.ns, name, data, subresources...), &v1alpha1.Azure{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Azure), err
}
