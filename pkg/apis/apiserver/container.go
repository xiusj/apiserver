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
	r "apiserver/pkg/router"
	"apiserver/pkg/storage/cache"
	httpUtil "apiserver/pkg/util/registry"

	"github.com/gorilla/mux"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/remotecommand"
)

// maximum number of lines loaded from the apiserver
var lineReadLimit int64 = 5000

// maximum number of bytes loaded from the apiserver
var byteReadLimit int64 = 500000

func GetContainers(request *http.Request) (string, interface{}) {
	// namespace := mux.Vars(request)["namespace"]
	pageCnt, _ := strconv.Atoi(request.FormValue("pageCnt"))
	pageNum, _ := strconv.Atoi(request.FormValue("pageNum"))
	containerName := request.FormValue("name")
	id, _ := strconv.ParseUint(mux.Vars(request)["id"], 10, 64)
	containers, total := apiserver.QueryContainers(containerName, pageCnt, pageNum, uint(id))
	return r.StatusOK, map[string]interface{}{"containers": containers, "total": total}
}

func RedeployContainer(request *http.Request) (string, interface{}) {
	id, _ := strconv.ParseUint(mux.Vars(request)["id"], 10, 64)
	container := apiserver.QueryContainerById(uint(id))
	svc := apiserver.QueryServiceById(container.ServiceId)
	namespace := mux.Vars(request)["namespace"]

	if !cache.ExsitResource(namespace, svc.Name, resource.ResourceKindDeployment) {
		return r.StatusNotFound, "service named " + svc.Name + ` does't exist`
	}

	pods, err := client.Client.GetPods(namespace, svc.Name)
	if err != nil {
		return r.StatusInternalServerError, err
	}
	for _, pod := range pods {
		if pod.Name == container.Name {
			if err = client.Client.DeleteResource(&pod); err != nil {
				return r.StatusInternalServerError, err
			}
		}
	}

	for _, c := range svc.Items {
		if container.Name == c.Name {
			apiserver.DeleteContainer(c)
		}
	}
	svc = apiserver.QueryServiceById(container.ServiceId)
	svc.Status = resource.AppRunning
	apiserver.UpdateService(svc)
	cache.ListResource()
	cache.UpdateAppStatus()

	return r.StatusCreated, "ok"
}

//GetContainerEvents return the pod's events ，default last one hour log
func GetContainerEvents(request *http.Request) (string, interface{}) {
	namespace := mux.Vars(request)["namespace"]
	containerName := mux.Vars(request)["name"]
	list, err := client.Client.GetEventsForContainer(namespace, containerName)
	if err != nil {
		return r.StatusInternalServerError, err
	}
	return r.StatusOK, map[string]interface{}{"events": list}
}

//GetContainerEvents return the pod's log，default the all log.
//it suport 1 hour 6 hour 1 day 1 week 1 month

func GetContainerLog(request *http.Request) (string, interface{}) {
	namespace := mux.Vars(request)["namespace"]
	podName := mux.Vars(request)["name"]
	sinceTimeSTR := request.FormValue("sinceTime")

	nowTime := time.Now()
	sinceTime := metav1.NewTime(nowTime)
	switch sinceTimeSTR {
	case "1h":
		sinceTime = metav1.NewTime(time.Unix(nowTime.Unix()-60*60, 0))
	case "6h":
		sinceTime = metav1.NewTime(time.Unix(nowTime.Unix()-6*60*60, 0))
	case "1d":
		sinceTime = metav1.NewTime(nowTime.AddDate(0, 0, -1))
	case "1w":
		sinceTime = metav1.NewTime(nowTime.AddDate(0, 0, -7))
	case "1m":
		sinceTime = metav1.NewTime(nowTime.AddDate(0, -1, 0))
	}
	containerName := cache.Store.PodCache.List[namespace][podName].Spec.Containers[0].Name
	logOptions := &v1.PodLogOptions{
		Container:  containerName,
		Follow:     false,
		Previous:   false,
		SinceTime:  &sinceTime,
		Timestamps: true,
		LimitBytes: &byteReadLimit,
		TailLines:  &lineReadLimit,
	}

	result, err := client.Client.GetLogForContainer(namespace, podName, logOptions)
	if err != nil {
		return r.StatusInternalServerError, err
	}
	return r.StatusOK, map[string]interface{}{"logs": result}
}

func GetContainerProcess(request *http.Request) (string, interface{}) {
	namespace := mux.Vars(request)["namespace"]
	name := mux.Vars(request)["name"]
	pod, exist := cache.Store.PodCache.List[namespace][name]
	if !exist {
		return r.StatusNotFound, fmt.Sprintf("container named [%v] doesn't exist", pod.Name)
	}

	containerID := strings.Replace(pod.Status.ContainerStatuses[0].ContainerID, "docker://", "", -1)
	tranport := httpUtil.GetHttpTransport(false)
	url := configz.GetString("apiserver", "cadvisor", "http://127.0.0.1:4194")
	client := &http.Client{Transport: tranport}
	url = fmt.Sprintf(url, pod.Status.HostIP)
	res, err := client.Get(url + `/api/v1.0/containers/kubepods`)
	if err != nil {
		return r.StatusInternalServerError, fmt.Sprintf("get process of container [%s] err:%v", pod.Name, err.Error())
	}

	var (
		cadvisorResult = apiserver.CadvisorResult{}
		processUrl     = ""
	)

	if err = json.NewDecoder(res.Body).Decode(&cadvisorResult); err != nil {
		return r.StatusInternalServerError, err
	}

	for k := range cadvisorResult.Subcontainers {
		cadvisorResult1 := apiserver.CadvisorResult{}
		subContainer := cadvisorResult.Subcontainers[k].Name
		res, err := client.Get(url + "/api/v1.0/containers" + subContainer)
		if err = json.NewDecoder(res.Body).Decode(&cadvisorResult1); err != nil {
			return r.StatusInternalServerError, err
		}
		defer res.Close()
		for k := range cadvisorResult1.Subcontainers {
			subContainer := cadvisorResult1.Subcontainers[k].Name
			if strings.Contains(subContainer, containerID) {
				processUrl = "/api/v2.0/ps" + subContainer
				goto PROCESS
			}
		}
	}

PROCESS:
	result, err := client.Get(url + processUrl)
	if err != nil {
		return r.StatusInternalServerError, fmt.Sprintf("get process of container [%s] err:%v", pod.Name, err.Error())
	}

	processes := []*apiserver.Process{}
	if err = json.NewDecoder(result.Body).Decode(&processes); err != nil {
		return r.StatusInternalServerError, err
	}
	defer result.Close()
	return r.StatusOK, map[string]interface{}{"processes": processes}
}

// Handles execute shell API call
func HandleExecShell(request *http.Request) (string, interface{}) {
	sessionId, err := genTerminalSessionId()
	if err != nil {
		return r.StatusInternalServerError, err
	}

	terminalSessions[sessionId] = TerminalSession{
		id:       sessionId,
		bound:    make(chan error),
		sizeChan: make(chan remotecommand.TerminalSize),
	}
	go WaitForTerminal(client.Client.K8sClient, client.Client.K8sConfig, request, sessionId)

	return r.StatusOK, map[string]interface{}{"session_id": sessionId}
}
