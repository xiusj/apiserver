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

	r "apiserver/pkg/router"

	"github.com/gorilla/mux"
)

func InstallApi(router *mux.Router) {
	//install apiserver's system api,include health check and get apiserver's version api
	r.RegisterHttpHandler(router, "/apiserver/health", "GET", Health)
	r.RegisterHttpHandler(router, "/apiserver/version", "GET", GetApiserverInfo)

	//install app's api handle
	r.RegisterHttpHandler(router, "/{namespace}/apps", "GET", GetApps)
	r.RegisterHttpHandler(router, "/{namespace}/apps/{id}", "GET", GetApp)
	r.RegisterHttpHandler(router, "/{namespace}/apps", "POST", CreateApp)
	r.RegisterHttpHandler(router, "/{namespace}/apps/{id}", "DELETE", DeleteApp)
	r.RegisterHttpHandler(router, "/{namespace}/apps/{id}/{verb}", "PATCH", StopOrStartOrRedeployApp)
	r.RegisterHttpHandler(router, "/{namespace}/apps/{id}/{verb}", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/{namespace}/apps", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/{namespace}/apps/", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/{namespace}/apps/{id}/", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/{namespace}/apps/{id}", "OPTIONS", Option)

	//install service's api handle
	r.RegisterHttpHandler(router, "/{namespace}/services", "GET", GetServices)
	r.RegisterHttpHandler(router, "/{namespace}/services/{id}", "GET", GetService)
	r.RegisterHttpHandler(router, "/{namespace}/services/{name}/events", "GET", GetServiceEvents)
	r.RegisterHttpHandler(router, "/{namespace}/services", "POST", CreateService)
	r.RegisterHttpHandler(router, "/{namespace}/services/{id}", "DELETE", DeleteService)
	r.RegisterHttpHandler(router, "/{namespace}/services/{id}/{verb}", "PUT", UpdateServiceConfig)
	r.RegisterHttpHandler(router, "/{namespace}/services/{id}/{verb}", "PATCH", StopOrStartOrRedployService)
	r.RegisterHttpHandler(router, "/{namespace}/services", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/{namespace}/services/{id}", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/{namespace}/services/{id}/", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/{namespace}/services/{id}/{verb}", "OPTIONS", Option)

	//install container's api handle
	r.RegisterHttpHandler(router, "/{namespace}/containers", "GET", GetContainers)
	r.RegisterHttpHandler(router, "/{namespace}/containers/{name}/events", "GET", GetContainerEvents)
	r.RegisterHttpHandler(router, "/{namespace}/containers/{id}", "PATCH", RedeployContainer)
	r.RegisterHttpHandler(router, "/{namespace}/containers", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/{namespace}/containers/", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/{namespace}/containers/{name}/events", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/{namespace}/containers/{name}/logs", "GET", GetContainerLog)
	r.RegisterHttpHandler(router, "/{namespace}/containers/{name}/processes", "GET", GetContainerProcess)
	r.RegisterHttpHandler(router, "/{namespace}/containers/{name}/exec", "GET", HandleExecShell)

	//install metrics's api handle
	r.RegisterHttpHandler(router, "/{namespace}/metrics/{name}/{metric}/{type}", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/{namespace}/metrics/{name}/{metric}/{type}", "GET", GetMetrics)
	// r.RegisterHttpHandler(router, "/{namespace}/pod/", "OPTIONS", Option)

	//install configMap's api handle
	r.RegisterHttpHandler(router, "/{namespace}/configs", "GET", GetConfigs)
	r.RegisterHttpHandler(router, "/{namespace}/configs", "POST", CreateConfig)
	r.RegisterHttpHandler(router, "/{namespace}/configs/{id}", "DELETE", DeleteConfig)
	r.RegisterHttpHandler(router, "/{namespace}/configs/{id}/items", "POST", CreateConfigItem)
	r.RegisterHttpHandler(router, "/{namespace}/configs/{id}/items/{itemId}", "DELETE", DeleteConfigItem)
	r.RegisterHttpHandler(router, "/{namespace}/configs", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/{namespace}/configs/{id}", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/{namespace}/configs/{id}/items", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/{namespace}/configs/{id}/items/{itemId}", "OPTIONS", Option)

	//install deploy's api handle
	r.RegisterHttpHandler(router, "/{namespace}/deploys", "POST", CreatDeploy)
	r.RegisterHttpHandler(router, "/{namespace}/deploys", "GET", QueryNoDeployedProject)
	r.RegisterHttpHandler(router, "/{namespace}/deploys", "OPTIONS", Option)

	//dashboard's count api
	r.RegisterHttpHandler(router, "/cluster/apps", "GET", GetAppCount)
	r.RegisterHttpHandler(router, "/cluster/services", "GET", GetServiceCount)
	r.RegisterHttpHandler(router, "/cluster/containers", "GET", GetContainerCount)
	r.RegisterHttpHandler(router, "/cluster/clusterInfo", "GET", GetCluterInfo)
	r.RegisterHttpHandler(router, "/cluster/clusterInfo/nodes", "GET", GetCluserNodes)
	r.RegisterHttpHandler(router, "/cluster/apps", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/cluster/services", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/cluster/containers", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/cluster/clusterInfo", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/cluster/clusterInfo/nodes", "OPTIONS", Option)
}

func Option(request *http.Request) (string, interface{}) {
	return r.StatusOK, "ok"
}
