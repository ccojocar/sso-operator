//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CookieSpec) DeepCopyInto(out *CookieSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CookieSpec.
func (in *CookieSpec) DeepCopy() *CookieSpec {
	if in == nil {
		return nil
	}
	out := new(CookieSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SSO) DeepCopyInto(out *SSO) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SSO.
func (in *SSO) DeepCopy() *SSO {
	if in == nil {
		return nil
	}
	out := new(SSO)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SSO) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SSOList) DeepCopyInto(out *SSOList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SSO, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SSOList.
func (in *SSOList) DeepCopy() *SSOList {
	if in == nil {
		return nil
	}
	out := new(SSOList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SSOList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SSOSpec) DeepCopyInto(out *SSOSpec) {
	*out = *in
	in.ProxyResources.DeepCopyInto(&out.ProxyResources)
	out.CookieSpec = in.CookieSpec
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SSOSpec.
func (in *SSOSpec) DeepCopy() *SSOSpec {
	if in == nil {
		return nil
	}
	out := new(SSOSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SSOStatus) DeepCopyInto(out *SSOStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SSOStatus.
func (in *SSOStatus) DeepCopy() *SSOStatus {
	if in == nil {
		return nil
	}
	out := new(SSOStatus)
	in.DeepCopyInto(out)
	return out
}
