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

	"apiserver/pkg/api/apiserver"
	"apiserver/pkg/client"
	"apiserver/pkg/configz"
	"apiserver/pkg/resource"
	"apiserver/pkg/resource/deployment"
	"apiserver/pkg/resource/service"
	"apiserver/pkg/resource/statefulset"
	r "apiserver/pkg/router"
	"apiserver/pkg/storage/cache"
	"apiserver/pkg/util/log"
	"apiserver/pkg/util/parseUtil"
	httpUtil "apiserver/pkg/util/registry"

	"github.com/gorilla/mux"
)

func GetApps(request *http.Request) (string, interface{}) {
	namespace := mux.Vars(request)["namespace"]
	pageCnt, _ := strconv.Atoi(request.FormValue("pageCnt"))
	pageNum, _ := strconv.Atoi(request.FormValue("pageNum"))
	appName := request.FormValue("name")
	apps, total := apiserver.QueryApps(namespace, appName, pageCnt, pageNum)
	return r.StatusOK, map[string]interface{}{"apps": apps, "total": total}
}

func GetApp(request *http.Request) (string, interface{}) {
	id, _ := strconv.ParseUint(mux.Vars(request)["id"], 10, 64)
	return r.StatusOK, apiserver.QueryAppById(uint(id))

}

func CreateApp(request *http.Request) (string, interface{}) {
	app, err := validateApp(request)
	image := app.Items[0].Image[0:strings.LastIndex(app.Items[0].Image, ":")]
	tag := app.Items[0].Image[strings.LastIndex(app.Items[0].Image, ":")+1:]
	serviceName := app.Items[0].Name
	if err != nil {
		go notify(serviceName, image, tag, "FAILED")
		return r.StatusBadRequest, err
	}

	if cache.ExsitResource(app.UserName, app.Items[0].Name, resource.ResourceKindService) {
		go notify(serviceName, image, tag, "FAILED")
		return r.StatusForbidden, "the service exist"
	}

	k8ssvc := service.NewService(app)
	svc, err := client.Client.CreateService(k8ssvc)
	if err != nil {
		go notify(serviceName, image, tag, "FAILED")
		return r.StatusInternalServerError, err
	}

	if app.Items[0].Type == 0 {
		k8sDeploy := deployment.NewDeployment(app)
		if err = client.Client.CreateResource(k8sDeploy); err != nil {
			if err = client.Client.DeleteResource(*svc); err != nil {
				go notify(serviceName, image, tag, "FAILED")
				return r.StatusInternalServerError, err
			}
			go notify(serviceName, image, tag, "FAILED")
			return r.StatusInternalServerError, err
		}
	}

	if app.Items[0].Type == 1 {
		k8sStatefulSet := statefulset.NewStatefulSet(app)
		if err = client.Client.CreateResource(k8sStatefulSet); err != nil {
			if err = client.Client.DeleteResource(*svc); err != nil {
				go notify(serviceName, image, tag, "FAILED")
				return r.StatusInternalServerError, err
			}
			go notify(serviceName, image, tag, "FAILED")
			return r.StatusInternalServerError, err
		}
	}

	external := fmt.Sprintf("http://%s:%v", configz.GetString("apiserver", "clusterNodes", "127.0.0.1"), svc.Spec.Ports[0].NodePort)
	app.External = external
	app.Items[0].External = external
	app.AppStatus = resource.AppBuilding

	apiserver.InsertApp(app)

	go notify(serviceName, image, tag, "SUCCESS")
	return r.StatusCreated, "ok"
}

func DeleteApp(request *http.Request) (string, interface{}) {
	id, _ := strconv.ParseUint(mux.Vars(request)["id"], 10, 64)
	namespace := mux.Vars(request)["namespace"]
	app := apiserver.QueryAppById(uint(id))

	for _, service := range app.Items {
		appName := service.Name

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
	}

	for _, svc := range app.Items {
		delete(cache.Store.ServiceCache.List[namespace], svc.Name)
		delete(cache.Store.DeploymentCache.List[namespace], svc.Name)
		for _, c := range svc.Items {
			delete(cache.Store.PodCache.List[namespace], c.Name)
		}
	}

	apiserver.DeleteApp(app)
	return r.StatusNoContent, "ok"
}

func StopOrStartOrRedeployApp(request *http.Request) (string, interface{}) {
	id, _ := strconv.ParseUint(mux.Vars(request)["id"], 10, 64)
	app := apiserver.QueryAppById(uint(id))

	for _, svc := range app.Items {
		appName := svc.Name
		namespace := mux.Vars(request)["namespace"]
		verb := mux.Vars(request)["verb"] //verb the action of app , start or stop

		if !cache.ExsitResource(namespace, appName, resource.ResourceKindDeployment) {
			return r.StatusNotFound, "service named " + appName + ` does't exist`
		}
		deploy := cache.Store.DeploymentCache.List[namespace][appName]
		if verb == "stop" {
			deploy.Spec.Replicas = parseUtil.IntToInt32Pointer(0)
			if err := client.Client.UpdateResouce(&deploy); err != nil {
				return r.StatusInternalServerError, err
			}

			app.AppStatus = resource.AppStop
			svc.Status = resource.AppStop
			apiserver.UpdateAppAndService(app)

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

			app.AppStatus = resource.AppRunning
			svc.Status = resource.AppRunning
			apiserver.UpdateAppOnly(app)
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
			app.AppStatus = resource.AppBuilding
			svc.Status = resource.AppBuilding
			apiserver.UpdateAppOnly(app)
			apiserver.UpdateServiceOnly(svc)

			for _, container := range svc.Items {
				delete(cache.Store.PodCache.List[namespace], container.Name)
				apiserver.DeleteContainer(container)
			}

		}
	}
	return r.StatusCreated, "ok"
}

func validateApp(request *http.Request) (*apiserver.App, error) {
	app := &apiserver.App{}
	if err := json.NewDecoder(request.Body).Decode(app); err != nil {
		return nil, err
	}
	return app, nil
}

func validateConfig(request *http.Request) (*apiserver.ServiceConfig, error) {
	cfg := &apiserver.ServiceConfig{}
	if err := json.NewDecoder(request.Body).Decode(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

//notify 部署反馈
//1. 会重配置文件中获取通知接口
//2. 根据部署的应用的镜像信息去获取对应的部署信息
//3. 将部署应用成功的信息发送
func notify(serviceName, image, tag, status string) {
	deploy := apiserver.QueryDeployByImage(image, tag)
	if deploy == nil {
		log.Noticef("this application [%v] who's image [%v:%v] doesn't need to notify", serviceName, image, tag)
		return
	}
	projectConfig := apiserver.QueryProjectConfigByID(deploy.Items[0].ProjectId)
	result := &apiserver.Result{ID: deploy.ID, CallbackResult: status, CallbackType: deploy.Type, Operator: projectConfig.Operator}
	resultItem := &apiserver.ResultItem{CurrentVersion: deploy.Items[0].Tag, ID: deploy.Items[0].ID, Status: status}
	result.Items = []*apiserver.ResultItem{resultItem}
	data, _ := json.Marshal(result)

	tranport := httpUtil.GetHttpTransport(false)
	url := configz.GetString("apiserver", "notifyUrl", "http://127.0.0.1")
	client := &http.Client{Transport: tranport}
	request, _ := http.NewRequest("PUT", url, strings.NewReader(string(data)))
	res, err := client.Do(request)
	if err != nil {
		log.Errorf("notify deploy message err:%v", err)
		return
	}
	if res.Status == "200" {
		log.Errorf("notify deploy message success")
	}
}
