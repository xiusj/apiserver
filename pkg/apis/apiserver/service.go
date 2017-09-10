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

package apiserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"apiserver/pkg/api/apiserver"
	"apiserver/pkg/client"
	"apiserver/pkg/configz"
	"apiserver/pkg/resource"
	"apiserver/pkg/resource/deployment"
	"apiserver/pkg/resource/service"
	r "apiserver/pkg/router"
	"apiserver/pkg/storage/cache"
	"apiserver/pkg/util/parseUtil"

	"github.com/gorilla/mux"
	"k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	res "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func GetServices(request *http.Request) (string, interface{}) {
	pageCnt := request.FormValue("pageCnt")
	pageNum := request.FormValue("pageNum")
	cnt, _ := strconv.Atoi(pageCnt)
	num, _ := strconv.Atoi(pageNum)
	serviceName := request.FormValue("name")
	appId, _ := strconv.ParseUint(request.FormValue("appId"), 10, 64)
	svcs, total := apiserver.QueryServices(serviceName, cnt, num, uint(appId))
	return r.StatusOK, map[string]interface{}{"services": svcs, "total": total}
}

func GetService(request *http.Request) (string, interface{}) {
	id, _ := strconv.ParseUint(mux.Vars(request)["id"], 10, 64)
	return r.StatusOK, map[string]interface{}{"service": apiserver.QueryServiceById(uint(id))}
}

func CreateService(request *http.Request) (string, interface{}) {
	reqservice, err := validateService(request)
	image := reqservice.Image[0:strings.LastIndex(reqservice.Image, ":")]
	tag := reqservice.Image[strings.LastIndex(reqservice.Image, ":")+1:]
	if err != nil {
		go notify(reqservice.Name, image, tag, "FAILED")
		return r.StatusBadRequest, err
	}

	app := apiserver.GetAppOnly(reqservice.AppId)
	app.Items = []*apiserver.Service{reqservice}
	if cache.ExsitResource(app.UserName, reqservice.Name, resource.ResourceKindService) {
		go notify(reqservice.Name, image, tag, "FAILED")
		return r.StatusForbidden, fmt.Sprintf("the service named [%v] exist", reqservice.Name)
	}

	k8ssvc := service.NewService(app)
	svc, err := client.Client.CreateService(k8ssvc)
	if err != nil {
		go notify(reqservice.Name, image, tag, "FAILED")
		return r.StatusInternalServerError, err
	}

	k8sDeploy := deployment.NewDeployment(app)
	if cache.ExsitResource(app.UserName, reqservice.Name, resource.ResourceKindDeployment) {
		go notify(reqservice.Name, image, tag, "FAILED")
		return r.StatusForbidden, "the deployment exist"
	}
	if err = client.Client.CreateResource(k8sDeploy); err != nil {
		if err = client.Client.DeleteResource(*svc); err != nil {
			go notify(reqservice.Name, image, tag, "FAILED")
			return r.StatusInternalServerError, err
		}
		go notify(reqservice.Name, image, tag, "FAILED")
		return r.StatusInternalServerError, err
	}
	external := fmt.Sprintf("http://%s:%v", configz.GetString("apiserver", "clusterNodes", "127.0.0.1"), svc.Spec.Ports[0].NodePort)
	app.External = external
	app.Items[0].External = external

	podList, err := client.Client.GetDeploymentPods(svc.Name, app.UserName)
	if err != nil {
		go notify(reqservice.Name, image, tag, "FAILED")
		return r.StatusInternalServerError, err
	}

	var cs []*apiserver.Container
	for _, pod := range podList {
		c := &apiserver.Container{Name: pod.ObjectMeta.Name, Image: app.Items[0].Image, Internal: pod.Status.PodIP}
		if pod.Status.Phase == "running" {
			c.Status = resource.AppRunning
		}

		cs = append(cs, c)
	}
	app.Items[0].Items = cs
	app.Items[0].AppName = app.Name
	apiserver.InsertService(app.Items[0])
	go notify(reqservice.Name, image, tag, "SUCCESS")
	return r.StatusCreated, "ok"
}

func DeleteService(request *http.Request) (string, interface{}) {
	id, _ := strconv.ParseUint(mux.Vars(request)["id"], 10, 64)
	svc := apiserver.QueryServiceById(uint(id))
	app := apiserver.GetAppOnly(svc.ID)
	appName := svc.Name
	namespace := app.UserName
	if !cache.ExsitResource(namespace, appName, resource.ResourceKindDeployment) {
		return r.StatusNotFound, "application named " + appName + ` does't exist`
	}
	if err := client.Client.DeleteResource(cache.Store.DeploymentCache.List[namespace][appName]); err != nil {
		return r.StatusInternalServerError, "delete application err: " + err.Error()
	}

	if !cache.ExsitResource(namespace, appName, resource.ResourceKindService) {
		return r.StatusNotFound, "application named " + appName + ` does't exist`
	}
	if err := client.Client.DeleteResource(cache.Store.ServiceCache.List[namespace][appName]); err != nil {
		return r.StatusInternalServerError, "delete application err: " + err.Error()
	}

	delete(cache.Store.ServiceCache.List[namespace], svc.Name)
	delete(cache.Store.DeploymentCache.List[namespace], svc.Name)
	for _, c := range svc.Items {
		delete(cache.Store.PodCache.List[namespace], c.Name)
	}

	apiserver.DeleteService(svc)
	return r.StatusOK, "ok"
}

func UpdateServiceConfig(request *http.Request) (string, interface{}) {
	id, _ := strconv.ParseUint(mux.Vars(request)["id"], 10, 64)
	svc := apiserver.QueryServiceById(uint(id))
	namespace := mux.Vars(request)["namespace"]
	verb := mux.Vars(request)["verb"] //verb the action of app , scale or expansion or roll

	if !cache.ExsitResource(namespace, svc.Name, resource.ResourceKindDeployment) {
		return r.StatusNotFound, "service named " + svc.Name + ` does't exist`
	}
	deploy := cache.Store.DeploymentCache.List[namespace][svc.Name]
	if verb == "scale" {
		scaleOption, err := validateScaleOption(request)
		if err != nil {
			return r.StatusBadRequest, err
		}
		deploy.Spec.Replicas = parseUtil.IntToInt32Pointer(scaleOption.ServiceInstanceCnt)
		svc.InstanceCount = scaleOption.ServiceInstanceCnt
		if err := client.Client.UpdateResouce(&deploy); err != nil {
			return r.StatusInternalServerError, "rolling updte application named " + svc.Name + ` failed`
		}
		for _, container := range svc.Items {
			delete(cache.Store.PodCache.List[namespace], container.Name)
			apiserver.DeleteContainer(container)
		}
		apiserver.UpdateService(svc)
	}

	if verb == "expansion" {
		expansionOption, err := validateExpansionOption(request)
		if err != nil {
			return r.StatusBadRequest, err
		}

		deploy.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceCPU:    res.MustParse(expansionOption.Cpu),    //TODO 根据前端传入的值做资源限制
				v1.ResourceMemory: res.MustParse(expansionOption.Memory), //TODO 根据前端传入的值做资源限制
			},
			Requests: v1.ResourceList{
				v1.ResourceCPU:    res.MustParse(expansionOption.Cpu),
				v1.ResourceMemory: res.MustParse(expansionOption.Memory),
			},
		}
		svc.Config.BaseConfig.Cpu = expansionOption.Cpu
		svc.Config.BaseConfig.Memory = expansionOption.Memory

		svc.Items[0].Config.BaseConfig.Cpu = expansionOption.Cpu
		svc.Items[0].Config.BaseConfig.Memory = expansionOption.Memory
		if err := client.Client.UpdateResouce(&deploy); err != nil {
			return r.StatusInternalServerError, "rolling updte application named " + svc.Name + ` failed`
		}
		apiserver.UpdateService(svc)
	}

	if verb == "roll" {
		volumes := []v1.Volume{}
		volumeMs := []v1.VolumeMount{}
		rollOption, err := validateRollOption(request)
		if err != nil {
			return r.StatusBadRequest, err
		}
		deploy.Spec.Template.Spec.Containers[0].Image = rollOption.Image
		deploy.Spec.Strategy = extensions.DeploymentStrategy{
			Type: extensions.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &extensions.RollingUpdateDeployment{
				MaxUnavailable: &intstr.IntOrString{Type: intstr.String, StrVal: "20%"},
				MaxSurge:       &intstr.IntOrString{Type: intstr.String, StrVal: "120%"},
			},
		}

		svc.Image = rollOption.Image
		svc.Items[0].Image = rollOption.Image
		if rollOption.Conifg != nil {
			items := []v1.KeyToPath{}
			for _, configMap := range rollOption.Conifg.ConfigMaps {
				items = append(items, v1.KeyToPath{Key: configMap.Name, Path: configMap.Name})
				containerPath := ""
				if configMap.ContainerPath[len(configMap.ContainerPath)-2:] == "/" {
					containerPath = configMap.ContainerPath + configMap.Name
				} else {
					containerPath = configMap.ContainerPath + "/" + configMap.Name
				}
				volumeMs = append(volumeMs, v1.VolumeMount{
					Name:      rollOption.Conifg.Name,
					MountPath: containerPath,
					SubPath:   configMap.Name,
				},
				)
			}
			volumes = append(volumes, v1.Volume{
				Name: rollOption.Conifg.Name,
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: rollOption.Conifg.Name},
						Items: items,
					},
				},
			})
		}
		deploy.Spec.Template.Spec.Volumes = volumes
		deploy.Spec.Template.Spec.Containers[0].VolumeMounts = volumeMs
		if err := client.Client.UpdateResouce(&deploy); err != nil {
			return r.StatusInternalServerError, "rolling updte application named " + svc.Name + ` failed`
		}
		apiserver.UpdateService(svc)
	}

	return r.StatusCreated, "ok"
}

func StopOrStartOrRedployService(request *http.Request) (string, interface{}) {
	id, _ := strconv.ParseUint(mux.Vars(request)["id"], 10, 64)
	svc := apiserver.QueryServiceById(uint(id))
	namespace := mux.Vars(request)["namespace"]
	verb := mux.Vars(request)["verb"] //verb the action of app , start or stop or redeploy

	if !cache.ExsitResource(namespace, svc.Name, resource.ResourceKindDeployment) {
		return r.StatusNotFound, "service named " + svc.Name + ` does't exist`
	}
	deploy := cache.Store.DeploymentCache.List[namespace][svc.Name]

	if verb == "stop" {
		deploy.Spec.Replicas = parseUtil.IntToInt32Pointer(0)
		if err := client.Client.UpdateResouce(&deploy); err != nil {
			return r.StatusInternalServerError, err
		}

		svc.Status = resource.AppStop
		apiserver.UpdateServiceOnly(svc)
		for _, container := range svc.Items {
			delete(cache.Store.PodCache.List[namespace], container.Name)
			apiserver.DeleteContainer(container)
		}
	}
	if verb == "start" {
		deploy.Spec.Replicas = parseUtil.IntToInt32Pointer(svc.InstanceCount)

		if err := client.Client.UpdateResouce(&deploy); err != nil {
			return r.StatusInternalServerError, err
		}

		svc.Status = resource.AppRunning
		apiserver.UpdateServiceOnly(svc)
		cache.ListResource()
		cache.UpdateAppStatus()
	}
	if verb == "redeploy" {
		pods, err := client.Client.GetPods(namespace, svc.Name)
		if err != nil {
			return r.StatusInternalServerError, err
		}
		for _, pod := range pods {
			if err = client.Client.DeleteResource(pod); err != nil {
				return r.StatusInternalServerError, err
			}
		}

		svc.Status = resource.AppBuilding
		apiserver.UpdateServiceOnly(svc)

		for _, container := range svc.Items {
			delete(cache.Store.PodCache.List[namespace], container.Name)
			apiserver.DeleteContainer(container)
		}

		cache.ListResource()
		cache.UpdateAppStatus()
	}
	return r.StatusCreated, "ok"
}

func validateService(request *http.Request) (*apiserver.Service, error) {
	svc := &apiserver.Service{}
	if err := json.NewDecoder(request.Body).Decode(svc); err != nil {
		return nil, err
	}
	return svc, nil
}

func validateScaleOption(request *http.Request) (*apiserver.ScaleOption, error) {
	option := &apiserver.ScaleOption{}
	if err := json.NewDecoder(request.Body).Decode(option); err != nil {
		return nil, err
	}
	return option, nil
}

func validateExpansionOption(request *http.Request) (*apiserver.ExpansionOption, error) {
	option := &apiserver.ExpansionOption{}
	if err := json.NewDecoder(request.Body).Decode(option); err != nil {
		return nil, err
	}
	return option, nil
}

func validateRollOption(request *http.Request) (*apiserver.RollOption, error) {
	option := &apiserver.RollOption{}
	if err := json.NewDecoder(request.Body).Decode(option); err != nil {
		return nil, err
	}
	return option, nil
}

func ChangeServiceStatus(svc *apiserver.Service, namespace string) error {
	time.Sleep(5 * time.Second)
	podList, err := client.Client.GetPods(namespace, svc.Name)
	if err != nil {
		return err
	}

	var containers []*apiserver.Container
	for _, pod := range podList {
		container := &apiserver.Container{Name: pod.ObjectMeta.Name, Image: svc.Image, Internal: pod.Status.PodIP, ServiceId: svc.ID}
		// if pod.Status.ContainerStatuses[0].Ready == true {
		// 	container.Status = resource.AppRunning
		// }
		containers = append(containers, container)

	}
	svc.Items = containers
	apiserver.UpdateService(svc)
	return nil
}

func GetServiceEvents(request *http.Request) (string, interface{}) {
	namespace := mux.Vars(request)["namespace"]
	containerName := mux.Vars(request)["name"]
	list, err := client.Client.GetEventsForContainer(namespace, containerName)
	if err != nil {
		return r.StatusInternalServerError, err
	}
	return r.StatusOK, map[string]interface{}{"events": list}
}
