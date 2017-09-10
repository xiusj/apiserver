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

package deployment

import (
	"apiserver/pkg/api/apiserver"
	"apiserver/pkg/resource"
	"apiserver/pkg/resource/pod"
	"apiserver/pkg/util/parseUtil"

	"k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewDeploymentSpce(app *apiserver.App) extensions.DeploymentSpec {
	return extensions.DeploymentSpec{
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
		Strategy: extensions.DeploymentStrategy{
			Type: extensions.RecreateDeploymentStrategyType,
			// RollingUpdate: extensions.RollingUpdateDeployment{},
		},
	}
}

func NewDeployment(app *apiserver.App) *extensions.Deployment {
	return &extensions.Deployment{
		TypeMeta:   resource.NewTypeMeta(resource.ResourceKindDeployment, "extensions/v1beta1"),
		ObjectMeta: resource.NewObjectMeta(app),
		Spec:       NewDeploymentSpce(app),
	}
}
