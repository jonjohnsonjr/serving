/*
Copyright 2018 The Knative Authors

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
	v1beta1 "github.com/knative/serving/pkg/apis/serving/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeConfigurations implements ConfigurationInterface
type FakeConfigurations struct {
	Fake *FakeServingV1beta1
	ns   string
}

var configurationsResource = schema.GroupVersionResource{Group: "serving.knative.dev", Version: "v1beta1", Resource: "configurations"}

var configurationsKind = schema.GroupVersionKind{Group: "serving.knative.dev", Version: "v1beta1", Kind: "Configuration"}

// Get takes name of the configuration, and returns the corresponding configuration object, and an error if there is any.
func (c *FakeConfigurations) Get(name string, options v1.GetOptions) (result *v1beta1.Configuration, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(configurationsResource, c.ns, name), &v1beta1.Configuration{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Configuration), err
}

// List takes label and field selectors, and returns the list of Configurations that match those selectors.
func (c *FakeConfigurations) List(opts v1.ListOptions) (result *v1beta1.ConfigurationList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(configurationsResource, configurationsKind, c.ns, opts), &v1beta1.ConfigurationList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.ConfigurationList{ListMeta: obj.(*v1beta1.ConfigurationList).ListMeta}
	for _, item := range obj.(*v1beta1.ConfigurationList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested configurations.
func (c *FakeConfigurations) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(configurationsResource, c.ns, opts))

}

// Create takes the representation of a configuration and creates it.  Returns the server's representation of the configuration, and an error, if there is any.
func (c *FakeConfigurations) Create(configuration *v1beta1.Configuration) (result *v1beta1.Configuration, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(configurationsResource, c.ns, configuration), &v1beta1.Configuration{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Configuration), err
}

// Update takes the representation of a configuration and updates it. Returns the server's representation of the configuration, and an error, if there is any.
func (c *FakeConfigurations) Update(configuration *v1beta1.Configuration) (result *v1beta1.Configuration, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(configurationsResource, c.ns, configuration), &v1beta1.Configuration{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Configuration), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeConfigurations) UpdateStatus(configuration *v1beta1.Configuration) (*v1beta1.Configuration, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(configurationsResource, "status", c.ns, configuration), &v1beta1.Configuration{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Configuration), err
}

// Delete takes name of the configuration and deletes it. Returns an error if one occurs.
func (c *FakeConfigurations) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(configurationsResource, c.ns, name), &v1beta1.Configuration{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeConfigurations) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(configurationsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1beta1.ConfigurationList{})
	return err
}

// Patch applies the patch and returns the patched configuration.
func (c *FakeConfigurations) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.Configuration, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(configurationsResource, c.ns, name, data, subresources...), &v1beta1.Configuration{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Configuration), err
}
