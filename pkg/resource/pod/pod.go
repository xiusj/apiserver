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

package pod

import (
	"apiserver/pkg/api/apiserver"
	res "apiserver/pkg/resource"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

//newPodSpec create k8s's PodSpec
func NewPodSpec(app *apiserver.App) v1.PodSpec {
	var (
		containerPorts       []v1.ContainerPort
		envs                 []v1.EnvVar
		volumes              []v1.Volume
		volumeMs             []v1.VolumeMount
		resourceRequirements v1.ResourceRequirements
		svc                  = app.Items[0]
	)
	if svc.Config != nil {
		ports := svc.Config.SuperConfig.Ports
		if len(ports) != 0 {
			for _, port := range ports {
				containerPorts = append(containerPorts, v1.ContainerPort{
					HostPort:      int32(port.ServicePort),
					ContainerPort: int32(port.ContainerPort),
					Protocol:      v1.Protocol(port.Protocol),
				})
			}
		}

		if len(svc.Config.SuperConfig.Envs) != 0 {
			for _, env := range svc.Config.SuperConfig.Envs {
				envs = append(envs, v1.EnvVar{Name: env.Key, Value: env.Val})
			}
		}

		if svc.Config.ConfigGroup != nil {
			items := []v1.KeyToPath{}
			for _, configMap := range svc.Config.ConfigGroup.ConfigMaps {
				items = append(items, v1.KeyToPath{Key: configMap.Name, Path: configMap.Name})
				// containerPath = configMap.ContainerPath
				containerPath := ""
				if configMap.ContainerPath[len(configMap.ContainerPath)-2:] == "/" {
					containerPath = configMap.ContainerPath + configMap.Name
				} else {
					containerPath = configMap.ContainerPath + "/" + configMap.Name
				}
				volumeMs = append(volumeMs, v1.VolumeMount{
					Name:      svc.Config.ConfigGroup.Name,
					MountPath: containerPath,
					SubPath:   configMap.Name,
				},
				)
			}
			volumes = append(volumes, v1.Volume{
				Name: svc.Config.ConfigGroup.Name,
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: svc.Config.ConfigGroup.Name},
						Items: items,
					},
				},
			})
		}

		if app.Items[0].Storage != nil {
			volumes = append(volumes, v1.Volume{
				Name: app.UserName + "_data",
				VolumeSource: v1.VolumeSource{
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
						ClaimName: app.UserName,
					},
				},
			},
			)
		}

		if svc.Config.BaseConfig != nil {
			resourceRequirements = v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(svc.Config.BaseConfig.Cpu),    //TODO 根据前端传入的值做资源限制
					v1.ResourceMemory: resource.MustParse(svc.Config.BaseConfig.Memory), //TODO 根据前端传入的值做资源限制
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(svc.Config.BaseConfig.Cpu),
					v1.ResourceMemory: resource.MustParse(svc.Config.BaseConfig.Memory),
				},
			}
		}
	}
	return v1.PodSpec{
		Volumes:       volumes,
		RestartPolicy: v1.RestartPolicyAlways,
		Containers: []v1.Container{
			v1.Container{
				Name:            svc.Name,
				Image:           svc.Image,
				ImagePullPolicy: v1.PullIfNotPresent,
				Resources:       resourceRequirements,
				Ports:           containerPorts,
				Env:             envs,
				VolumeMounts:    volumeMs,
			},
		},
	}
}

func NewPod(app *apiserver.App) *v1.Pod {
	return &v1.Pod{
		TypeMeta:   res.NewTypeMeta(res.ResourceKindPod, "v1"),
		ObjectMeta: res.NewObjectMeta(app),
		Spec:       NewPodSpec(app),
	}
}
