// Copyright 2018-2020 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !privileged_tests

package k8s

import (
	"time"

	"github.com/cilium/cilium/pkg/annotation"
	"github.com/cilium/cilium/pkg/checker"
	fakeDatapath "github.com/cilium/cilium/pkg/datapath/fake"
	v2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	slim_corev1 "github.com/cilium/cilium/pkg/k8s/slim/k8s/apis/core/v1"
	slim_metav1 "github.com/cilium/cilium/pkg/k8s/slim/k8s/apis/meta/v1"
	"github.com/cilium/cilium/pkg/k8s/types"
	"github.com/cilium/cilium/pkg/labels"
	"github.com/cilium/cilium/pkg/policy/api"

	. "gopkg.in/check.v1"
	"k8s.io/api/core/v1"
	core_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func (s *K8sSuite) Test_EqualV2CNP(c *C) {
	type args struct {
		o1 *types.SlimCNP
		o2 *types.SlimCNP
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "CNP with the same name",
			args: args{
				o1: &types.SlimCNP{
					CiliumNetworkPolicy: &v2.CiliumNetworkPolicy{
						ObjectMeta: metav1.ObjectMeta{
							Name: "rule1",
						},
					},
				},
				o2: &types.SlimCNP{
					CiliumNetworkPolicy: &v2.CiliumNetworkPolicy{
						ObjectMeta: metav1.ObjectMeta{
							Name: "rule1",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "CNP with the different spec",
			args: args{
				o1: &types.SlimCNP{
					CiliumNetworkPolicy: &v2.CiliumNetworkPolicy{
						ObjectMeta: metav1.ObjectMeta{
							Name: "rule1",
						},
						Spec: &api.Rule{
							EndpointSelector: api.NewESFromLabels(labels.NewLabel("foo", "bar", "k8s")),
						},
					},
				},
				o2: &types.SlimCNP{
					CiliumNetworkPolicy: &v2.CiliumNetworkPolicy{
						ObjectMeta: metav1.ObjectMeta{
							Name: "rule1",
						},
						Spec: nil,
					},
				},
			},
			want: false,
		},
		{
			name: "CNP with the same spec",
			args: args{
				o1: &types.SlimCNP{
					CiliumNetworkPolicy: &v2.CiliumNetworkPolicy{
						ObjectMeta: metav1.ObjectMeta{
							Name: "rule1",
						},
						Spec: &api.Rule{},
					},
				},
				o2: &types.SlimCNP{
					CiliumNetworkPolicy: &v2.CiliumNetworkPolicy{
						ObjectMeta: metav1.ObjectMeta{
							Name: "rule1",
						},
						Spec: &api.Rule{},
					},
				},
			},
			want: true,
		},
		{
			name: "CNP with different last applied annotations. The are ignored so they should be equal",
			args: args{
				o1: &types.SlimCNP{
					CiliumNetworkPolicy: &v2.CiliumNetworkPolicy{
						ObjectMeta: metav1.ObjectMeta{
							Name: "rule1",
							Annotations: map[string]string{
								core_v1.LastAppliedConfigAnnotation: "foo",
							},
						},
						Spec: &api.Rule{},
					},
				},
				o2: &types.SlimCNP{
					CiliumNetworkPolicy: &v2.CiliumNetworkPolicy{
						ObjectMeta: metav1.ObjectMeta{
							Name: "rule1",
							Annotations: map[string]string{
								core_v1.LastAppliedConfigAnnotation: "bar",
							},
						},
						Spec: &api.Rule{},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		got := EqualV2CNP(tt.args.o1, tt.args.o2)
		c.Assert(got, Equals, tt.want, Commentf("Test Name: %s", tt.name))
	}
}

func (s *K8sSuite) Test_EqualV1Endpoints(c *C) {
	type args struct {
		o1 *slim_corev1.Endpoints
		o2 *slim_corev1.Endpoints
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "EPs with the same name",
			args: args{
				o1: &slim_corev1.Endpoints{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "rule1",
					},
				},
				o2: &slim_corev1.Endpoints{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "rule1",
					},
				},
			},
			want: true,
		},
		{
			name: "EPs with the different spec",
			args: args{
				o1: &slim_corev1.Endpoints{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "rule1",
					},
					Subsets: []slim_corev1.EndpointSubset{
						{
							Addresses: []slim_corev1.EndpointAddress{
								{
									IP: "172.0.0.1",
								},
							},
						},
					},
				},
				o2: &slim_corev1.Endpoints{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "rule1",
					},
				},
			},
			want: false,
		},
		{
			name: "EPs with the same spec",
			args: args{
				o1: &slim_corev1.Endpoints{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "rule1",
					},
					Subsets: []slim_corev1.EndpointSubset{
						{
							Addresses: []slim_corev1.EndpointAddress{
								{
									IP: "172.0.0.1",
								},
							},
						},
					},
				},
				o2: &slim_corev1.Endpoints{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "rule1",
					},
					Subsets: []slim_corev1.EndpointSubset{
						{
							Addresses: []slim_corev1.EndpointAddress{
								{
									IP: "172.0.0.1",
								},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "EPs with the same spec (multiple IPs)",
			args: args{
				o1: &slim_corev1.Endpoints{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "rule1",
					},
					Subsets: []slim_corev1.EndpointSubset{
						{
							Addresses: []slim_corev1.EndpointAddress{
								{
									IP: "172.0.0.1",
								},
								{
									IP: "172.0.0.2",
								},
							},
						},
					},
				},
				o2: &slim_corev1.Endpoints{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "rule1",
					},
					Subsets: []slim_corev1.EndpointSubset{
						{
							Addresses: []slim_corev1.EndpointAddress{
								{
									IP: "172.0.0.1",
								},
								{
									IP: "172.0.0.2",
								},
							},
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		got := EqualV1Endpoints(tt.args.o1, tt.args.o2)
		c.Assert(got, Equals, tt.want, Commentf("Test Name: %s", tt.name))
	}
}

func (s *K8sSuite) Test_EqualV1Pod(c *C) {
	type args struct {
		o1 *slim_corev1.Pod
		o2 *slim_corev1.Pod
	}
	tests := []struct {
		name string
		args args
		want bool
	}{

		{
			name: "Pods with the same name",
			args: args{
				o1: &slim_corev1.Pod{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "pod1",
					},
				},
				o2: &slim_corev1.Pod{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "pod1",
					},
				},
			},
			want: true,
		},
		{
			name: "Pods with the different spec",
			args: args{
				o1: &slim_corev1.Pod{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "pod1",
					},
					Status: slim_corev1.PodStatus{
						HostIP: "127.0.0.1",
						PodIPs: []slim_corev1.PodIP{
							{
								IP: "127.0.0.2",
							},
						},
					},
				},
				o2: &slim_corev1.Pod{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "pod1",
					},
					Status: slim_corev1.PodStatus{
						HostIP: "127.0.0.1",
						PodIPs: []slim_corev1.PodIP{
							{
								IP: "127.0.0.1",
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "Pods with the same spec",
			args: args{
				o1: &slim_corev1.Pod{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "pod1",
					},
					Status: slim_corev1.PodStatus{
						HostIP: "127.0.0.1",
						PodIPs: []slim_corev1.PodIP{
							{
								IP: "127.0.0.2",
							},
						},
					},
				},
				o2: &slim_corev1.Pod{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "pod1",
					},
					Status: slim_corev1.PodStatus{
						HostIP: "127.0.0.1",
						PodIPs: []slim_corev1.PodIP{
							{
								IP: "127.0.0.2",
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Pods with the same spec but different labels",
			args: args{
				o1: &slim_corev1.Pod{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "pod1",
						Labels: map[string]string{
							"foo": "bar",
						},
					},
					Status: slim_corev1.PodStatus{
						HostIP: "127.0.0.1",
						PodIPs: []slim_corev1.PodIP{
							{
								IP: "127.0.0.2",
							},
						},
					},
				},
				o2: &slim_corev1.Pod{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "pod1",
					},
					Status: slim_corev1.PodStatus{
						HostIP: "127.0.0.1",
						PodIPs: []slim_corev1.PodIP{
							{
								IP: "127.0.0.2",
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "Pods with the same spec and same labels",
			args: args{
				o1: &slim_corev1.Pod{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "pod1",
						Labels: map[string]string{
							"foo": "bar",
						},
					},
					Status: slim_corev1.PodStatus{
						HostIP: "127.0.0.1",
						PodIPs: []slim_corev1.PodIP{
							{
								IP: "127.0.0.2",
							},
						},
					},
				},
				o2: &slim_corev1.Pod{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "pod1",
						Labels: map[string]string{
							"foo": "bar",
						},
					},
					Status: slim_corev1.PodStatus{
						HostIP: "127.0.0.1",
						PodIPs: []slim_corev1.PodIP{
							{
								IP: "127.0.0.2",
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Pods with differing proxy-visibility annotations",
			args: args{
				o1: &slim_corev1.Pod{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "pod1",
						Labels: map[string]string{
							"foo": "bar",
						},
					},
					Status: slim_corev1.PodStatus{
						HostIP: "127.0.0.1",
						PodIPs: []slim_corev1.PodIP{
							{
								IP: "127.0.0.2",
							},
						},
					},
				},
				o2: &slim_corev1.Pod{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "pod1",
						Labels: map[string]string{
							"foo": "bar",
						},
						Annotations: map[string]string{
							annotation.ProxyVisibility: "80/HTTP",
						},
					},
					Status: slim_corev1.PodStatus{
						HostIP: "127.0.0.1",
						PodIPs: []slim_corev1.PodIP{
							{
								IP: "127.0.0.2",
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "Pods with irrelevant annotations",
			args: args{
				o1: &slim_corev1.Pod{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "pod1",
						Labels: map[string]string{
							"foo": "bar",
						},
					},
					Status: slim_corev1.PodStatus{
						HostIP: "127.0.0.1",
						PodIPs: []slim_corev1.PodIP{
							{
								IP: "127.0.0.2",
							},
						},
					},
				},
				o2: &slim_corev1.Pod{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "pod1",
						Labels: map[string]string{
							"foo": "bar",
						},
						Annotations: map[string]string{
							"useless": "80/HTTP",
						},
					},
					Status: slim_corev1.PodStatus{
						HostIP: "127.0.0.1",
						PodIPs: []slim_corev1.PodIP{
							{
								IP: "127.0.0.2",
							},
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		got := EqualV1Pod(tt.args.o1, tt.args.o2)
		c.Assert(got, Equals, tt.want, Commentf("Test Name: %s", tt.name))
	}
}

func (s *K8sSuite) Test_EqualV1Node(c *C) {
	type args struct {
		o1 *slim_corev1.Node
		o2 *slim_corev1.Node
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Nodes with the same name",
			args: args{
				o1: &slim_corev1.Node{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Node1",
					},
				},
				o2: &slim_corev1.Node{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Node1",
					},
				},
			},
			want: true,
		},
		{
			name: "Nodes with the different names",
			args: args{
				o1: &slim_corev1.Node{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Node1",
					},
				},
				o2: &slim_corev1.Node{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Node2",
					},
				},
			},
			want: false,
		},
		{
			name: "Nodes with the different spec should return true as we don't care about the spec",
			args: args{
				o1: &slim_corev1.Node{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Node1",
					},
					Spec: slim_corev1.NodeSpec{
						PodCIDR: "192.168.0.0/10",
					},
				},
				o2: &slim_corev1.Node{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Node1",
					},
					Spec: slim_corev1.NodeSpec{
						PodCIDR: "127.0.0.1/10",
					},
				},
			},
			want: true,
		},
		{
			name: "Nodes with the same annotations",
			args: args{
				o1: &slim_corev1.Node{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Node1",
						Annotations: map[string]string{
							annotation.CiliumHostIP: "127.0.0.1",
						},
					},
				},
				o2: &slim_corev1.Node{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Node1",
						Annotations: map[string]string{
							annotation.CiliumHostIP: "127.0.0.1",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Nodes with the different annotations",
			args: args{
				o1: &slim_corev1.Node{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Node1",
						Annotations: map[string]string{
							annotation.CiliumHostIP: "127.0.0.1",
						},
					},
				},
				o2: &slim_corev1.Node{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Node1",
						Annotations: map[string]string{
							annotation.CiliumHostIP: "127.0.0.2",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "Nodes with the same annotations and different specs should return true because he don't care about the spec",
			args: args{
				o1: &slim_corev1.Node{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Node1",
						Annotations: map[string]string{
							annotation.CiliumHostIP: "127.0.0.1",
						},
					},
					Spec: slim_corev1.NodeSpec{
						PodCIDR: "192.168.0.0/10",
					},
				},
				o2: &slim_corev1.Node{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Node1",
						Annotations: map[string]string{
							annotation.CiliumHostIP: "127.0.0.1",
						},
					},
					Spec: slim_corev1.NodeSpec{
						PodCIDR: "127.0.0.1/10",
					},
				},
			},
			want: true,
		},
		{
			name: "Nodes with the same taints and different specs should return true because he don't care about the spec",
			args: args{
				o1: &slim_corev1.Node{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Node1",
					},
					Spec: slim_corev1.NodeSpec{
						PodCIDR: "192.168.0.0/10",
						Taints: []slim_corev1.Taint{
							{
								Key:    "key",
								Value:  "value",
								Effect: "no-effect",
							},
						},
					},
				},
				o2: &slim_corev1.Node{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Node1",
					},
					Spec: slim_corev1.NodeSpec{
						PodCIDR: "127.0.0.1/10",
						Taints: []slim_corev1.Taint{
							{
								Key:    "key",
								Value:  "value",
								Effect: "no-effect",
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Nodes with the same taints and different specs should return true because he don't care about the spec",
			args: args{
				o1: &slim_corev1.Node{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Node1",
					},
					Spec: slim_corev1.NodeSpec{
						PodCIDR: "192.168.0.0/10",
						Taints: []slim_corev1.Taint{
							{
								Key:       "key",
								Value:     "value",
								Effect:    "no-effect",
								TimeAdded: func() *slim_metav1.Time { return &slim_metav1.Time{Time: time.Unix(1, 1)} }(),
							},
						},
					},
				},
				o2: &slim_corev1.Node{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Node1",
					},
					Spec: slim_corev1.NodeSpec{
						PodCIDR: "127.0.0.1/10",
						Taints: []slim_corev1.Taint{
							{
								Key:       "key",
								Value:     "value",
								Effect:    "no-effect",
								TimeAdded: func() *slim_metav1.Time { return &slim_metav1.Time{Time: time.Unix(1, 1)} }(),
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Nodes with the different taints and different specs should return true because he don't care about the spec",
			args: args{
				o1: &slim_corev1.Node{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Node1",
					},
					Spec: slim_corev1.NodeSpec{
						PodCIDR: "192.168.0.0/10",
						Taints: []slim_corev1.Taint{
							{
								Key:    "key",
								Value:  "value",
								Effect: "no-effect",
							},
						},
					},
				},
				o2: &slim_corev1.Node{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Node1",
					}, Spec: slim_corev1.NodeSpec{
						PodCIDR: "127.0.0.1/10",
						Taints: []slim_corev1.Taint{
							{
								Key:       "key",
								Value:     "value",
								Effect:    "no-effect",
								TimeAdded: func() *slim_metav1.Time { return &slim_metav1.Time{Time: time.Unix(1, 1)} }(),
							},
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		got := EqualV1Node(tt.args.o1, tt.args.o2)
		c.Assert(got, Equals, tt.want, Commentf("Test Name: %s", tt.name))
	}
}

func (s *K8sSuite) Test_EqualV1Namespace(c *C) {
	type args struct {
		o1 *slim_corev1.Namespace
		o2 *slim_corev1.Namespace
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Namespaces with the same name",
			args: args{
				o1: &slim_corev1.Namespace{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Namespace1",
					},
				},
				o2: &slim_corev1.Namespace{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Namespace1",
					},
				},
			},
			want: true,
		},
		{
			name: "Namespaces with the different names",
			args: args{
				o1: &slim_corev1.Namespace{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Namespace1",
					},
				},
				o2: &slim_corev1.Namespace{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Namespace2",
					},
				},
			},
			want: false,
		},
		{
			name: "Namespaces with the same labels",
			args: args{
				o1: &slim_corev1.Namespace{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Namespace1",
						Labels: map[string]string{
							"prod": "true",
						},
					},
				},
				o2: &slim_corev1.Namespace{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Namespace1",
						Labels: map[string]string{
							"prod": "true",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Namespaces with the different labels",
			args: args{
				o1: &slim_corev1.Namespace{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Namespace1",
						Labels: map[string]string{
							"prod": "true",
						},
					},
				},
				o2: &slim_corev1.Namespace{
					ObjectMeta: slim_metav1.ObjectMeta{
						Name: "Namespace1",
						Labels: map[string]string{
							"prod": "false",
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		got := EqualV1Namespace(tt.args.o1, tt.args.o2)
		c.Assert(got, Equals, tt.want, Commentf("Test Name: %s", tt.name))
	}
}

func (s *K8sSuite) Test_EqualV1Service(c *C) {
	type args struct {
		o1 *slim_corev1.Service
		o2 *slim_corev1.Service
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Service with different annotations",
			args: args{
				o1: &slim_corev1.Service{
					ObjectMeta: slim_metav1.ObjectMeta{
						Annotations: map[string]string{},
					},
				},
				o2: &slim_corev1.Service{
					ObjectMeta: slim_metav1.ObjectMeta{
						Annotations: map[string]string{
							"io.cilium/shared-service": "true",
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		got := EqualV1Services(tt.args.o1, tt.args.o2, fakeDatapath.NewNodeAddressing())
		c.Assert(got, Equals, tt.want, Commentf("Test Name: %s", tt.name))
	}
}

func (s *K8sSuite) Test_ConvertToK8sService(c *C) {
	type args struct {
		obj interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "normal conversion",
			args: args{
				obj: &v1.Service{},
			},
			want: &slim_corev1.Service{},
		},
		{
			name: "delete final state unknown conversion",
			args: args{
				obj: cache.DeletedFinalStateUnknown{
					Key: "foo",
					Obj: &v1.Service{},
				},
			},
			want: cache.DeletedFinalStateUnknown{
				Key: "foo",
				Obj: &slim_corev1.Service{},
			},
		},
		{
			name: "unknown object type in delete final state unknown conversion",
			args: args{
				obj: cache.DeletedFinalStateUnknown{
					Key: "foo",
					Obj: 100,
				},
			},
			want: cache.DeletedFinalStateUnknown{
				Key: "foo",
				Obj: 100,
			},
		},
		{
			name: "unknown object type in conversion",
			args: args{
				obj: 100,
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		got := ConvertToK8sService(tt.args.obj)
		c.Assert(got, checker.DeepEquals, tt.want, Commentf("Test Name: %s", tt.name))
	}
}

func (s *K8sSuite) Test_ConvertToCNPWithStatus(c *C) {
	type args struct {
		obj interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "normal conversion",
			args: args{
				obj: &v2.CiliumNetworkPolicy{},
			},
			want: &types.SlimCNP{
				CiliumNetworkPolicy: &v2.CiliumNetworkPolicy{},
			},
		},
		{
			name: "delete final state unknown conversion",
			args: args{
				obj: cache.DeletedFinalStateUnknown{
					Key: "foo",
					Obj: &v2.CiliumNetworkPolicy{},
				},
			},
			want: cache.DeletedFinalStateUnknown{
				Key: "foo",
				Obj: &types.SlimCNP{
					CiliumNetworkPolicy: &v2.CiliumNetworkPolicy{},
				},
			},
		},
		{
			name: "unknown object type in delete final state unknown conversion",
			args: args{
				obj: cache.DeletedFinalStateUnknown{
					Key: "foo",
					Obj: 100,
				},
			},
			want: cache.DeletedFinalStateUnknown{
				Key: "foo",
				Obj: 100,
			},
		},
		{
			name: "unknown object type in conversion",
			args: args{
				obj: 100,
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		got := ConvertToCNPWithStatus(tt.args.obj)
		c.Assert(got, checker.DeepEquals, tt.want, Commentf("Test Name: %s", tt.name))
	}
}

func (s *K8sSuite) Test_ConvertToCNP(c *C) {
	type args struct {
		obj interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "normal conversion",
			args: args{
				obj: &v2.CiliumNetworkPolicy{},
			},
			want: &types.SlimCNP{
				CiliumNetworkPolicy: &v2.CiliumNetworkPolicy{},
			},
		},
		{
			name: "delete final state unknown conversion",
			args: args{
				obj: cache.DeletedFinalStateUnknown{
					Key: "foo",
					Obj: &v2.CiliumNetworkPolicy{},
				},
			},
			want: cache.DeletedFinalStateUnknown{
				Key: "foo",
				Obj: &types.SlimCNP{
					CiliumNetworkPolicy: &v2.CiliumNetworkPolicy{},
				},
			},
		},
		{
			name: "unknown object type in delete final state unknown conversion",
			args: args{
				obj: cache.DeletedFinalStateUnknown{
					Key: "foo",
					Obj: 100,
				},
			},
			want: cache.DeletedFinalStateUnknown{
				Key: "foo",
				Obj: 100,
			},
		},
		{
			name: "unknown object type in conversion",
			args: args{
				obj: 100,
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		got := ConvertToCNP(tt.args.obj)
		c.Assert(got, checker.DeepEquals, tt.want, Commentf("Test Name: %s", tt.name))
	}
}

func (s *K8sSuite) Test_ConvertToNode(c *C) {
	type args struct {
		obj interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "normal conversion",
			args: args{
				obj: &v1.Node{},
			},
			want: &slim_corev1.Node{},
		},
		{
			name: "delete final state unknown conversion",
			args: args{
				obj: cache.DeletedFinalStateUnknown{
					Key: "foo",
					Obj: &v1.Node{},
				},
			},
			want: cache.DeletedFinalStateUnknown{
				Key: "foo",
				Obj: &slim_corev1.Node{},
			},
		},
		{
			name: "unknown object type in delete final state unknown conversion",
			args: args{
				obj: cache.DeletedFinalStateUnknown{
					Key: "foo",
					Obj: 100,
				},
			},
			want: cache.DeletedFinalStateUnknown{
				Key: "foo",
				Obj: 100,
			},
		},
		{
			name: "unknown object type in conversion",
			args: args{
				obj: 100,
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		got := ConvertToNode(tt.args.obj)
		c.Assert(got, checker.DeepEquals, tt.want, Commentf("Test Name: %s", tt.name))
	}
}

func (s *K8sSuite) Test_ConvertToCiliumNode(c *C) {
	type args struct {
		obj interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "normal conversion",
			args: args{
				obj: &v2.CiliumNode{},
			},
			want: &v2.CiliumNode{},
		},
		{
			name: "delete final state unknown conversion",
			args: args{
				obj: cache.DeletedFinalStateUnknown{
					Key: "foo",
					Obj: &v2.CiliumNode{},
				},
			},
			want: cache.DeletedFinalStateUnknown{
				Key: "foo",
				Obj: &v2.CiliumNode{},
			},
		},
		{
			name: "unknown object type in delete final state unknown conversion",
			args: args{
				obj: cache.DeletedFinalStateUnknown{
					Key: "foo",
					Obj: 100,
				},
			},
			want: cache.DeletedFinalStateUnknown{
				Key: "foo",
				Obj: 100,
			},
		},
		{
			name: "unknown object type in conversion",
			args: args{
				obj: 100,
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		got := ConvertToCiliumNode(tt.args.obj)
		c.Assert(got, checker.DeepEquals, tt.want, Commentf("Test Name: %s", tt.name))
	}
}

func (s *K8sSuite) Test_ConvertToCiliumEndpoint(c *C) {
	type args struct {
		obj interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "normal conversion",
			args: args{
				obj: &v2.CiliumEndpoint{},
			},
			want: &types.CiliumEndpoint{
				Encryption: &v2.EncryptionSpec{},
			},
		},
		{
			name: "delete final state unknown conversion",
			args: args{
				obj: cache.DeletedFinalStateUnknown{
					Key: "foo",
					Obj: &v2.CiliumEndpoint{},
				},
			},
			want: cache.DeletedFinalStateUnknown{
				Key: "foo",
				Obj: &types.CiliumEndpoint{
					Encryption: &v2.EncryptionSpec{},
				},
			},
		},
		{
			name: "unknown object type in delete final state unknown conversion",
			args: args{
				obj: cache.DeletedFinalStateUnknown{
					Key: "foo",
					Obj: 100,
				},
			},
			want: cache.DeletedFinalStateUnknown{
				Key: "foo",
				Obj: 100,
			},
		},
		{
			name: "unknown object type in conversion",
			args: args{
				obj: 100,
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		got := ConvertToCiliumEndpoint(tt.args.obj)
		c.Assert(got, checker.DeepEquals, tt.want, Commentf("Test Name: %s", tt.name))
	}
}

func (s *K8sSuite) Test_AnnotationsEqual(c *C) {
	irrelevantAnnoKey := "foo"
	irrelevantAnnoVal := "bar"

	relevantAnnoKey := annotation.ProxyVisibility
	relevantAnnoVal1 := "<Ingress/80/TCP/HTTP>"
	relevantAnnoVal2 := "<Ingress/80/TCP/HTTP>,<Egress/80/TCP/HTTP>"

	// Empty returns true.
	c.Assert(AnnotationsEqual(nil, map[string]string{}, map[string]string{}), Equals, true)

	c.Assert(AnnotationsEqual(nil,
		map[string]string{
			irrelevantAnnoKey: irrelevantAnnoVal,
			relevantAnnoKey:   relevantAnnoVal1,
		}, map[string]string{
			irrelevantAnnoKey: irrelevantAnnoVal,
			relevantAnnoKey:   relevantAnnoVal2,
		}), Equals, true)

	// If the relevant annotation isn't in either map, return true.
	c.Assert(AnnotationsEqual([]string{relevantAnnoKey},
		map[string]string{
			irrelevantAnnoKey: irrelevantAnnoVal,
		}, map[string]string{
			irrelevantAnnoKey: irrelevantAnnoVal,
		}), Equals, true)

	c.Assert(AnnotationsEqual([]string{relevantAnnoKey},
		map[string]string{
			relevantAnnoKey: relevantAnnoVal1,
		}, map[string]string{
			relevantAnnoKey: relevantAnnoVal2,
		}), Equals, false)

}
