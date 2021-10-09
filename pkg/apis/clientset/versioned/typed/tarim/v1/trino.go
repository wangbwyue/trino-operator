/*


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
// Code generated by main. DO NOT EDIT.

package v1

import (
	"context"
	"time"
	v1 "trino-operator/apis/tarim/v1"
	scheme "trino-operator/pkg/apis/clientset/versioned/scheme"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// TrinosGetter has a method to return a TrinoInterface.
// A group's client should implement this interface.
type TrinosGetter interface {
	Trinos(namespace string) TrinoInterface
}

// TrinoInterface has methods to work with Trino resources.
type TrinoInterface interface {
	Create(ctx context.Context, trino *v1.Trino, opts metav1.CreateOptions) (*v1.Trino, error)
	Update(ctx context.Context, trino *v1.Trino, opts metav1.UpdateOptions) (*v1.Trino, error)
	UpdateStatus(ctx context.Context, trino *v1.Trino, opts metav1.UpdateOptions) (*v1.Trino, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Trino, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.TrinoList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Trino, err error)
	TrinoExpansion
}

// trinos implements TrinoInterface
type trinos struct {
	client rest.Interface
	ns     string
}

// newTrinos returns a Trinos
func newTrinos(c *TarimV1Client, namespace string) *trinos {
	return &trinos{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the trino, and returns the corresponding trino object, and an error if there is any.
func (c *trinos) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.Trino, err error) {
	result = &v1.Trino{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("trinos").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Trinos that match those selectors.
func (c *trinos) List(ctx context.Context, opts metav1.ListOptions) (result *v1.TrinoList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.TrinoList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("trinos").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested trinos.
func (c *trinos) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("trinos").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a trino and creates it.  Returns the server's representation of the trino, and an error, if there is any.
func (c *trinos) Create(ctx context.Context, trino *v1.Trino, opts metav1.CreateOptions) (result *v1.Trino, err error) {
	result = &v1.Trino{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("trinos").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(trino).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a trino and updates it. Returns the server's representation of the trino, and an error, if there is any.
func (c *trinos) Update(ctx context.Context, trino *v1.Trino, opts metav1.UpdateOptions) (result *v1.Trino, err error) {
	result = &v1.Trino{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("trinos").
		Name(trino.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(trino).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *trinos) UpdateStatus(ctx context.Context, trino *v1.Trino, opts metav1.UpdateOptions) (result *v1.Trino, err error) {
	result = &v1.Trino{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("trinos").
		Name(trino.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(trino).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the trino and deletes it. Returns an error if one occurs.
func (c *trinos) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("trinos").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *trinos) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("trinos").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched trino.
func (c *trinos) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Trino, err error) {
	result = &v1.Trino{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("trinos").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
