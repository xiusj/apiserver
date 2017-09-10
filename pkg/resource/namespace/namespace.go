// Copyright © 2017 huang jia <449264675@qq.com>
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

package namespace

import (
	"apiserver/pkg/api/apiserver"
	"apiserver/pkg/resource"

	"k8s.io/client-go/pkg/api/v1"
)

func NewNamespaceSpec(svc *apiserver.Service) v1.NamespaceSpec {
	return v1.NamespaceSpec{
		Finalizers: []v1.FinalizerName{v1.FinalizerKubernetes},
	}
}

func NewNamespace(svc *apiserver.Service) *v1.Namespace {
	return &v1.Namespace{
		TypeMeta:   resource.NewTypeMeta(resource.ResourceKindNamespace, "v1"),
		ObjectMeta: resource.NewObjectMeta(svc),
		Spec:       NewNamespaceSpec(svc),
	}
}
