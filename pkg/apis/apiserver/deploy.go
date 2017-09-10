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
	"time"

	"apiserver/pkg/api/apiserver"
	k8s "apiserver/pkg/client"
	"apiserver/pkg/configz"
	"apiserver/pkg/resource/configMap"
	r "apiserver/pkg/router"
	httpUtil "apiserver/pkg/util/registry"

	"github.com/gorilla/mux"
)

//CreatDeploy create deploy record
//1. get the request date of project
//2. get the config of the project by step 1 's project data
//3. insert the config and project data to the db ,at the same time create the k8s's resoure of configMap
func CreatDeploy(request *http.Request) (string, interface{}) {
	namespace := mux.Vars(request)["namespace"]
	deploys, err := validateDeploy(request)
	if err != nil {
		return r.StatusInternalServerError, err
	}

	tranport := httpUtil.GetHttpTransport(false)
	url := configz.GetString("apiserver", "getConfigUrl", "http://localhost:8080/projects/%s/configs")
	client := &http.Client{Transport: tranport}

	for _, deploy := range deploys {
		if err = apiserver.InsertDeploy(deploy); err != nil {
			return r.StatusInternalServerError, err
		}
		for _, item := range deploy.Items {
			projectConfigOptions := []*apiserver.ProjectConfigOption{}
			url = fmt.Sprintf(url, item.ProjectId)
			res, err := client.Get(url)
			if err != nil {
				return r.StatusInternalServerError, fmt.Sprintf("get project [%s] config err:%v", item.ProjectName, err.Error())
			}
			if err = json.NewDecoder(res.Body).Decode(&projectConfigOptions); err != nil {
				return r.StatusInternalServerError, err
			}
			configMaps := []*apiserver.ConfigMap{}
			for _, projectConfigOption := range projectConfigOptions {
				createAt, _ := time.Parse("2006-01-02 15:04:05", projectConfigOption.CreateAt)
				updateAt, _ := time.Parse("2006-01-02 15:04:05", projectConfigOption.UpdateAt)
				projectConfig := &apiserver.ProjectConfig{
					ProjectId: projectConfigOption.ProjectId,
					Key:       projectConfigOption.Key,
					Val:       projectConfigOption.Val,
					Type:      projectConfigOption.Type,
					CreateAt:  createAt,
					UpdateAt:  updateAt,
					Operator:  projectConfigOption.Operator,
				}
				if err = apiserver.InsertProjectConfig(projectConfig); err != nil {
					return r.StatusInternalServerError, err
				}
				configMaps = append(configMaps, &apiserver.ConfigMap{Name: projectConfigOption.Key, Content: projectConfigOption.Val})
			}

			configGroup := &apiserver.ConfigGroup{
				Name:       item.ProjectName + "_" + item.Tag,
				Namespace:  namespace,
				ConfigMaps: configMaps,
			}
			k8sConfigMap := configMap.NewConfigMapByConfig(configGroup)
			if err = k8s.Client.CreateResource(&k8sConfigMap); err != nil {
				return r.StatusInternalServerError, err
			}
		}
	}

	return r.StatusCreated, "ok"
}

func QueryNoDeployedProject(request *http.Request) (string, interface{}) {
	items, total := apiserver.QueryNoDeployItem()
	return r.StatusOK, map[string]interface{}{"items": items, "total": total}
}

func validateDeploy(request *http.Request) ([]*apiserver.Deploy, error) {
	deploys := []*apiserver.Deploy{}
	if err := json.NewDecoder(request.Body).Decode(&deploys); err != nil {
		return nil, err
	}
	return deploys, nil
}
