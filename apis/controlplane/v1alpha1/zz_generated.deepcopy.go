//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2021.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/api/v1beta1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IdentityRefSpec) DeepCopyInto(out *IdentityRefSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IdentityRefSpec.
func (in *IdentityRefSpec) DeepCopy() *IdentityRefSpec {
	if in == nil {
		return nil
	}
	out := new(IdentityRefSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KopsControlPlane) DeepCopyInto(out *KopsControlPlane) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KopsControlPlane.
func (in *KopsControlPlane) DeepCopy() *KopsControlPlane {
	if in == nil {
		return nil
	}
	out := new(KopsControlPlane)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KopsControlPlane) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KopsControlPlaneList) DeepCopyInto(out *KopsControlPlaneList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]KopsControlPlane, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KopsControlPlaneList.
func (in *KopsControlPlaneList) DeepCopy() *KopsControlPlaneList {
	if in == nil {
		return nil
	}
	out := new(KopsControlPlaneList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KopsControlPlaneList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KopsControlPlaneSpec) DeepCopyInto(out *KopsControlPlaneSpec) {
	*out = *in
	out.IdentityRef = in.IdentityRef
	in.KopsClusterSpec.DeepCopyInto(&out.KopsClusterSpec)
	if in.KopsSecret != nil {
		in, out := &in.KopsSecret, &out.KopsSecret
		*out = new(v1.ObjectReference)
		**out = **in
	}
	out.SpotInst = in.SpotInst
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KopsControlPlaneSpec.
func (in *KopsControlPlaneSpec) DeepCopy() *KopsControlPlaneSpec {
	if in == nil {
		return nil
	}
	out := new(KopsControlPlaneSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KopsControlPlaneStatus) DeepCopyInto(out *KopsControlPlaneStatus) {
	*out = *in
	if in.FailureMessage != nil {
		in, out := &in.FailureMessage, &out.FailureMessage
		*out = new(string)
		**out = **in
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make(v1beta1.Conditions, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Secrets != nil {
		in, out := &in.Secrets, &out.Secrets
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KopsControlPlaneStatus.
func (in *KopsControlPlaneStatus) DeepCopy() *KopsControlPlaneStatus {
	if in == nil {
		return nil
	}
	out := new(KopsControlPlaneStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpotInstSpec) DeepCopyInto(out *SpotInstSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpotInstSpec.
func (in *SpotInstSpec) DeepCopy() *SpotInstSpec {
	if in == nil {
		return nil
	}
	out := new(SpotInstSpec)
	in.DeepCopyInto(out)
	return out
}
