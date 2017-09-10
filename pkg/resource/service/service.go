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

package service

import (
	"apiserver/pkg/api/apiserver"
	"apiserver/pkg/resource"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewServiceSpec(app *apiserver.App) v1.ServiceSpec {
	var ports []v1.ServicePort
	svc := app.Items[0]
	if svc.Config.SuperConfig != nil {
		for _, port := range svc.Config.SuperConfig.Ports {
			ports = append(ports, v1.ServicePort{Port: int32(port.ServicePort), TargetPort: intstr.FromInt(port.ContainerPort), Protocol: v1.Protocol(port.Protocol)})
		}
	}
	return v1.ServiceSpec{
		Ports:    ports,
		Selector: map[string]string{"name": svc.Name},
		Type:     v1.ServiceTypeNodePort,
		// LoadBalancerIP: svc.LoadbalanceIp,
	}
}

func NewService(app *apiserver.App) *v1.Service {
	return &v1.Service{
		TypeMeta:   resource.NewTypeMeta(resource.ResourceKindDeployment, "extensions/v1beta1"),
		ObjectMeta: resource.NewObjectMeta(app),
		Spec:       NewServiceSpec(app),
	}
}
