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
package v1alpha1

import (
	v1alpha1 "github.com/knative/build-pipeline/pkg/apis/pipeline/v1alpha1"
	scheme "github.com/knative/build-pipeline/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ResourcesGetter has a method to return a ResourceInterface.
// A group's client should implement this interface.
type ResourcesGetter interface {
	Resources(namespace string) ResourceInterface
}

// ResourceInterface has methods to work with Resource resources.
type ResourceInterface interface {
	Create(*v1alpha1.Resource) (*v1alpha1.Resource, error)
	Update(*v1alpha1.Resource) (*v1alpha1.Resource, error)
	UpdateStatus(*v1alpha1.Resource) (*v1alpha1.Resource, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Resource, error)
	List(opts v1.ListOptions) (*v1alpha1.ResourceList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Resource, err error)
	ResourceExpansion
}

// resources implements ResourceInterface
type resources struct {
	client rest.Interface
	ns     string
}

// newResources returns a Resources
func newResources(c *PipelineV1alpha1Client, namespace string) *resources {
	return &resources{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the resource, and returns the corresponding resource object, and an error if there is any.
func (c *resources) Get(name string, options v1.GetOptions) (result *v1alpha1.Resource, err error) {
	result = &v1alpha1.Resource{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("resources").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Resources that match those selectors.
func (c *resources) List(opts v1.ListOptions) (result *v1alpha1.ResourceList, err error) {
	result = &v1alpha1.ResourceList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("resources").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested resources.
func (c *resources) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("resources").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a resource and creates it.  Returns the server's representation of the resource, and an error, if there is any.
func (c *resources) Create(resource *v1alpha1.Resource) (result *v1alpha1.Resource, err error) {
	result = &v1alpha1.Resource{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("resources").
		Body(resource).
		Do().
		Into(result)
	return
}

// Update takes the representation of a resource and updates it. Returns the server's representation of the resource, and an error, if there is any.
func (c *resources) Update(resource *v1alpha1.Resource) (result *v1alpha1.Resource, err error) {
	result = &v1alpha1.Resource{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("resources").
		Name(resource.Name).
		Body(resource).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *resources) UpdateStatus(resource *v1alpha1.Resource) (result *v1alpha1.Resource, err error) {
	result = &v1alpha1.Resource{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("resources").
		Name(resource.Name).
		SubResource("status").
		Body(resource).
		Do().
		Into(result)
	return
}

// Delete takes name of the resource and deletes it. Returns an error if one occurs.
func (c *resources) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("resources").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *resources) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("resources").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched resource.
func (c *resources) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Resource, err error) {
	result = &v1alpha1.Resource{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("resources").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
