//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2022.

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

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaInstance) DeepCopyInto(out *KafkaInstance) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaInstance.
func (in *KafkaInstance) DeepCopy() *KafkaInstance {
	if in == nil {
		return nil
	}
	out := new(KafkaInstance)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KafkaInstance) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaInstanceList) DeepCopyInto(out *KafkaInstanceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]KafkaInstance, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaInstanceList.
func (in *KafkaInstanceList) DeepCopy() *KafkaInstanceList {
	if in == nil {
		return nil
	}
	out := new(KafkaInstanceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KafkaInstanceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaInstanceSpec) DeepCopyInto(out *KafkaInstanceSpec) {
	*out = *in
	if in.ReauthenticationEnabled != nil {
		in, out := &in.ReauthenticationEnabled, &out.ReauthenticationEnabled
		*out = new(bool)
		**out = **in
	}
	if in.MultiAz != nil {
		in, out := &in.MultiAz, &out.MultiAz
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaInstanceSpec.
func (in *KafkaInstanceSpec) DeepCopy() *KafkaInstanceSpec {
	if in == nil {
		return nil
	}
	out := new(KafkaInstanceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaInstanceStatus) DeepCopyInto(out *KafkaInstanceStatus) {
	*out = *in
	in.CreatedAt.DeepCopyInto(&out.CreatedAt)
	in.UpdatedAt.DeepCopyInto(&out.UpdatedAt)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaInstanceStatus.
func (in *KafkaInstanceStatus) DeepCopy() *KafkaInstanceStatus {
	if in == nil {
		return nil
	}
	out := new(KafkaInstanceStatus)
	in.DeepCopyInto(out)
	return out
}
