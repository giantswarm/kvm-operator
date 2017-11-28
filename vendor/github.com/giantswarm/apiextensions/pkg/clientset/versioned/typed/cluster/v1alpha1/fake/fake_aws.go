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

// FakeAWSs implements AWSInterface
type FakeAWSs struct {
	Fake *FakeClusterV1alpha1
	ns   string
}

var awssResource = schema.GroupVersionResource{Group: "cluster.giantswarm.io", Version: "v1alpha1", Resource: "awss"}

var awssKind = schema.GroupVersionKind{Group: "cluster.giantswarm.io", Version: "v1alpha1", Kind: "AWS"}

// Get takes name of the aWS, and returns the corresponding aWS object, and an error if there is any.
func (c *FakeAWSs) Get(name string, options v1.GetOptions) (result *v1alpha1.AWS, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(awssResource, c.ns, name), &v1alpha1.AWS{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AWS), err
}

// List takes label and field selectors, and returns the list of AWSs that match those selectors.
func (c *FakeAWSs) List(opts v1.ListOptions) (result *v1alpha1.AWSList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(awssResource, awssKind, c.ns, opts), &v1alpha1.AWSList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.AWSList{}
	for _, item := range obj.(*v1alpha1.AWSList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested aWSs.
func (c *FakeAWSs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(awssResource, c.ns, opts))

}

// Create takes the representation of a aWS and creates it.  Returns the server's representation of the aWS, and an error, if there is any.
func (c *FakeAWSs) Create(aWS *v1alpha1.AWS) (result *v1alpha1.AWS, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(awssResource, c.ns, aWS), &v1alpha1.AWS{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AWS), err
}

// Update takes the representation of a aWS and updates it. Returns the server's representation of the aWS, and an error, if there is any.
func (c *FakeAWSs) Update(aWS *v1alpha1.AWS) (result *v1alpha1.AWS, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(awssResource, c.ns, aWS), &v1alpha1.AWS{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AWS), err
}

// Delete takes name of the aWS and deletes it. Returns an error if one occurs.
func (c *FakeAWSs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(awssResource, c.ns, name), &v1alpha1.AWS{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeAWSs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(awssResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.AWSList{})
	return err
}

// Patch applies the patch and returns the patched aWS.
func (c *FakeAWSs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.AWS, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(awssResource, c.ns, name, data, subresources...), &v1alpha1.AWS{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AWS), err
}
