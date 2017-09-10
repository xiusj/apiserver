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

package statefulset

import (
	"apiserver/pkg/api/apiserver"
	"apiserver/pkg/resource"
	"apiserver/pkg/resource/pod"
	"apiserver/pkg/util/parseUtil"

	betav1 "k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewStatefulSetSpec(app *apiserver.App) betav1.StatefulSetSpec {
	return betav1.StatefulSetSpec{
		Replicas: parseUtil.IntToInt32Pointer(app.Items[0].InstanceCount),
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"name": app.Items[0].Name,
			},
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: resource.NewObjectMeta(app),
			Spec:       pod.NewPodSpec(app),
		},
		ServiceName: app.Items[0].Name,
	}
}

func NewStatefulSetStatus(app *apiserver.App) betav1.StatefulSetStatus {
	return betav1.StatefulSetStatus{}
}

func NewStatefulSet(app *apiserver.App) *betav1.StatefulSet {
	return &betav1.StatefulSet{
		TypeMeta:   resource.NewTypeMeta("StatefulSet", "apps/v1beta1"),
		ObjectMeta: resource.NewObjectMeta(app),
		Spec:       NewStatefulSetSpec(app),
		Status:     NewStatefulSetStatus(app),
	}
}
