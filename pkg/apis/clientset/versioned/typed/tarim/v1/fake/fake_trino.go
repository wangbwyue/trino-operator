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
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"
	tarimv1 "trino-operator/apis/tarim/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeTrinos implements TrinoInterface
type FakeTrinos struct {
	Fake *FakeTarimV1
	ns   string
}

var trinosResource = schema.GroupVersionResource{Group: "tarim.deepexi.com", Version: "v1", Resource: "trinos"}

var trinosKind = schema.GroupVersionKind{Group: "tarim.deepexi.com", Version: "v1", Kind: "Trino"}

// Get takes name of the trino, and returns the corresponding trino object, and an error if there is any.
func (c *FakeTrinos) Get(ctx context.Context, name string, options v1.GetOptions) (result *tarimv1.Trino, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(trinosResource, c.ns, name), &tarimv1.Trino{})

	if obj == nil {
		return nil, err
	}
	return obj.(*tarimv1.Trino), err
}

// List takes label and field selectors, and returns the list of Trinos that match those selectors.
func (c *FakeTrinos) List(ctx context.Context, opts v1.ListOptions) (result *tarimv1.TrinoList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(trinosResource, trinosKind, c.ns, opts), &tarimv1.TrinoList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &tarimv1.TrinoList{ListMeta: obj.(*tarimv1.TrinoList).ListMeta}
	for _, item := range obj.(*tarimv1.TrinoList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested trinos.
func (c *FakeTrinos) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(trinosResource, c.ns, opts))

}

// Create takes the representation of a trino and creates it.  Returns the server's representation of the trino, and an error, if there is any.
func (c *FakeTrinos) Create(ctx context.Context, trino *tarimv1.Trino, opts v1.CreateOptions) (result *tarimv1.Trino, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(trinosResource, c.ns, trino), &tarimv1.Trino{})

	if obj == nil {
		return nil, err
	}
	return obj.(*tarimv1.Trino), err
}

// Update takes the representation of a trino and updates it. Returns the server's representation of the trino, and an error, if there is any.
func (c *FakeTrinos) Update(ctx context.Context, trino *tarimv1.Trino, opts v1.UpdateOptions) (result *tarimv1.Trino, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(trinosResource, c.ns, trino), &tarimv1.Trino{})

	if obj == nil {
		return nil, err
	}
	return obj.(*tarimv1.Trino), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeTrinos) UpdateStatus(ctx context.Context, trino *tarimv1.Trino, opts v1.UpdateOptions) (*tarimv1.Trino, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(trinosResource, "status", c.ns, trino), &tarimv1.Trino{})

	if obj == nil {
		return nil, err
	}
	return obj.(*tarimv1.Trino), err
}

// Delete takes name of the trino and deletes it. Returns an error if one occurs.
func (c *FakeTrinos) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(trinosResource, c.ns, name), &tarimv1.Trino{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeTrinos) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(trinosResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &tarimv1.TrinoList{})
	return err
}

// Patch applies the patch and returns the patched trino.
func (c *FakeTrinos) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *tarimv1.Trino, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(trinosResource, c.ns, name, pt, data, subresources...), &tarimv1.Trino{})

	if obj == nil {
		return nil, err
	}
	return obj.(*tarimv1.Trino), err
}
