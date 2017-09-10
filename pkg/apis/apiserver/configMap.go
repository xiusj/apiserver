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
	"net/http"
	"strconv"

	"apiserver/pkg/api/apiserver"
	"apiserver/pkg/client"
	"apiserver/pkg/resource/configMap"
	r "apiserver/pkg/router"

	"github.com/gorilla/mux"
)

func GetConfigs(request *http.Request) (string, interface{}) {
	pageCnt, _ := strconv.Atoi(request.FormValue("pageCnt"))
	pageNum, _ := strconv.Atoi(request.FormValue("pageNum"))
	configName := request.FormValue("name")
	configs, total := apiserver.QueryConfigs(configName, pageCnt, pageNum)
	return r.StatusOK, map[string]interface{}{"configs": configs, "total": total}
}

func CreateConfig(request *http.Request) (string, interface{}) {
	config, err := validate(request)
	if err != nil {
		return r.StatusInternalServerError, err
	}
	namespace := mux.Vars(request)["namespace"]
	config.Namespace = namespace
	cfgMap := configMap.NewConfigMapByConfig(config)
	if err = client.Client.CreateResource(&cfgMap); err != nil {
		return r.StatusInternalServerError, err
	}

	apiserver.InsertConfig(config)

	return r.StatusCreated, "ok"
}

func DeleteConfig(request *http.Request) (string, interface{}) {
	id, _ := strconv.ParseUint(mux.Vars(request)["id"], 10, 64)
	cfg := apiserver.QueryConfigById(uint(id))
	cfgMap := configMap.NewConfigMapByConfig(cfg)
	if err := client.Client.DeleteResource(cfgMap); err != nil {
		return r.StatusInternalServerError, err
	}
	apiserver.DeleteConfig(uint(id))
	return r.StatusOK, "ok"
}

func DeleteConfigItem(request *http.Request) (string, interface{}) {
	itemid, _ := strconv.ParseUint(mux.Vars(request)["itemId"], 10, 64)
	apiserver.DeleteConfigItem(uint(itemid))

	id, _ := strconv.ParseUint(mux.Vars(request)["id"], 10, 64)
	cfg := apiserver.QueryConfigById(uint(id))
	cfgMap := configMap.NewConfigMapByConfig(cfg)
	if err := client.Client.UpdateResouce(&cfgMap); err != nil {
		return r.StatusInternalServerError, err
	}
	return r.StatusOK, "ok"
}

func CreateConfigItem(request *http.Request) (string, interface{}) {
	id, _ := strconv.ParseUint(mux.Vars(request)["id"], 10, 64)
	cfgMap, err := validateConfigItem(request)
	if err != nil {
		return r.StatusInternalServerError, err
	}
	cfgMap.ConfigGroupId = uint(id)
	apiserver.InsertConfigItem(cfgMap)

	itemid, _ := strconv.ParseUint(mux.Vars(request)["itemId"], 10, 64)
	apiserver.DeleteConfigItem(uint(itemid))

	cfg := apiserver.QueryConfigById(uint(id))
	cfgk8s := configMap.NewConfigMapByConfig(cfg)
	if err := client.Client.UpdateResouce(&cfgk8s); err != nil {
		return r.StatusInternalServerError, err
	}

	return r.StatusCreated, "ok"
}

func validate(request *http.Request) (*apiserver.ConfigGroup, error) {
	config := &apiserver.ConfigGroup{}
	if err := json.NewDecoder(request.Body).Decode(config); err != nil {
		return nil, err
	}
	return config, nil
}

func validateConfigItem(request *http.Request) (*apiserver.ConfigMap, error) {
	config := &apiserver.ConfigMap{}
	if err := json.NewDecoder(request.Body).Decode(config); err != nil {
		return nil, err
	}
	return config, nil
}
