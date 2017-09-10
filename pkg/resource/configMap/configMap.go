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

package configMap

import (
	"apiserver/pkg/api/apiserver"
	"apiserver/pkg/resource"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewConfigMap(app *apiserver.App) *v1.ConfigMap {

	data := map[string]string{}

	if app.Items[0].Config.ConfigGroup != nil {
		for _, configMap := range app.Items[0].Config.ConfigGroup.ConfigMaps {
			data[configMap.Name] = configMap.Content
		}
	}

	return &v1.ConfigMap{
		TypeMeta:   resource.NewTypeMeta(resource.ResourceKindConfigMap, "v1"),
		ObjectMeta: resource.NewObjectMeta(app),
		Data:       data,
	}
}

func NewConfigMapByService(svc *apiserver.Service, namespace string) *v1.ConfigMap {
	data := map[string]string{}

	if svc.Config.ConfigGroup != nil {
		for _, configMap := range svc.Config.ConfigGroup.ConfigMaps {
			data[configMap.Name] = configMap.Content
		}
	}

	return &v1.ConfigMap{
		TypeMeta: resource.NewTypeMeta(resource.ResourceKindConfigMap, "v1"),
		ObjectMeta: metav1.ObjectMeta{
			Name:        svc.Name,
			Namespace:   namespace,
			Labels:      map[string]string{"name": svc.Name},
			Annotations: map[string]string{"name": svc.Name},
		},
		Data: data,
	}
}

func NewConfigMapByConfig(c *apiserver.ConfigGroup) v1.ConfigMap {
	datas := map[string]string{}
	for _, cfg := range c.ConfigMaps {
		datas[cfg.Name] = cfg.Content
	}

	return v1.ConfigMap{
		TypeMeta: resource.NewTypeMeta(resource.ResourceKindConfigMap, "v1"),
		ObjectMeta: metav1.ObjectMeta{
			Name:        c.Name,
			Namespace:   c.Namespace,
			Labels:      map[string]string{"name": c.Name},
			Annotations: map[string]string{"name": c.Name},
		},
		Data: datas,
	}
}
