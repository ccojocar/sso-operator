// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	jenkinsiov1 "github.com/jenkins-x/sso-operator/pkg/apis/jenkins.io/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeSSOs implements SSOInterface
type FakeSSOs struct {
	Fake *FakeJenkinsV1
	ns   string
}

var ssosResource = schema.GroupVersionResource{Group: "jenkins.io", Version: "v1", Resource: "ssos"}

var ssosKind = schema.GroupVersionKind{Group: "jenkins.io", Version: "v1", Kind: "SSO"}

// Get takes name of the sSO, and returns the corresponding sSO object, and an error if there is any.
func (c *FakeSSOs) Get(name string, options v1.GetOptions) (result *jenkinsiov1.SSO, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(ssosResource, c.ns, name), &jenkinsiov1.SSO{})

	if obj == nil {
		return nil, err
	}
	return obj.(*jenkinsiov1.SSO), err
}

// List takes label and field selectors, and returns the list of SSOs that match those selectors.
func (c *FakeSSOs) List(opts v1.ListOptions) (result *jenkinsiov1.SSOList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(ssosResource, ssosKind, c.ns, opts), &jenkinsiov1.SSOList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &jenkinsiov1.SSOList{ListMeta: obj.(*jenkinsiov1.SSOList).ListMeta}
	for _, item := range obj.(*jenkinsiov1.SSOList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested sSOs.
func (c *FakeSSOs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(ssosResource, c.ns, opts))

}

// Create takes the representation of a sSO and creates it.  Returns the server's representation of the sSO, and an error, if there is any.
func (c *FakeSSOs) Create(sSO *jenkinsiov1.SSO) (result *jenkinsiov1.SSO, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(ssosResource, c.ns, sSO), &jenkinsiov1.SSO{})

	if obj == nil {
		return nil, err
	}
	return obj.(*jenkinsiov1.SSO), err
}

// Update takes the representation of a sSO and updates it. Returns the server's representation of the sSO, and an error, if there is any.
func (c *FakeSSOs) Update(sSO *jenkinsiov1.SSO) (result *jenkinsiov1.SSO, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(ssosResource, c.ns, sSO), &jenkinsiov1.SSO{})

	if obj == nil {
		return nil, err
	}
	return obj.(*jenkinsiov1.SSO), err
}

// Delete takes name of the sSO and deletes it. Returns an error if one occurs.
func (c *FakeSSOs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(ssosResource, c.ns, name), &jenkinsiov1.SSO{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSSOs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(ssosResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &jenkinsiov1.SSOList{})
	return err
}

// Patch applies the patch and returns the patched sSO.
func (c *FakeSSOs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *jenkinsiov1.SSO, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(ssosResource, c.ns, name, data, subresources...), &jenkinsiov1.SSO{})

	if obj == nil {
		return nil, err
	}
	return obj.(*jenkinsiov1.SSO), err
}