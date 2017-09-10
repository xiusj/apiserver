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
	"net/http"

	"apiserver/pkg/api/apiserver"
	"apiserver/pkg/configz"
	r "apiserver/pkg/router"
)

func Health(request *http.Request) (string, interface{}) {
	return r.StatusOK, "apiserver is healthy"
}

func GetApiserverInfo(request *http.Request) (string, interface{}) {
	return r.StatusOK, map[string]interface{}{"componentName": configz.GetString("apiserver", "componentName", "apiserver"), "version": configz.GetString("apiserver", "version", "v1.0")}
}

func GetAppCount(request *http.Request) (string, interface{}) {
	result, err := apiserver.CountApp()
	if err != nil {
		return r.StatusInternalServerError, err
	}
	return r.StatusOK, result
}

func GetServiceCount(request *http.Request) (string, interface{}) {
	result, err := apiserver.CountService()
	if err != nil {
		return r.StatusInternalServerError, err
	}
	return r.StatusOK, result
}

func GetContainerCount(request *http.Request) (string, interface{}) {
	result, err := apiserver.CountContainer()
	if err != nil {
		return r.StatusInternalServerError, err
	}
	return r.StatusOK, result
}
