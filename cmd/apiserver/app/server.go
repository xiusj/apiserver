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

package app

import (
	"fmt"
	"net/http"

	"apiserver/pkg/apis/apiserver"
	"apiserver/pkg/componentconfig"
	"apiserver/pkg/configz"
	"apiserver/pkg/storage/cache"
	"apiserver/pkg/util/log"

	"github.com/gorilla/mux"
)

type Apiserver struct {
	*componentconfig.ApiserverConfig
}

func NewApiServer() *Apiserver {
	return &Apiserver{
		ApiserverConfig: &componentconfig.ApiserverConfig{
			HttpAddr: configz.GetString("apiserver", "httpAddr", "0.0.0.0"),
			HttpPort: configz.MustInt("apiserver", "httpPort", 9090),
			RpcAddr:  configz.GetString("apiserver", "rpcAddr", "0.0.0.0"),
			RpcPort:  configz.MustInt("apiserver", "rpcPort", 7070),
		},
	}
}

func Run(server *Apiserver) error {
	root := mux.NewRouter()
	api := root.PathPrefix("/api/v1").Subrouter()
	installApiGroup(api)
	http.Handle("/", root)
	log.Infof("starting apiserver and listen on : %v", fmt.Sprintf("%v:%v", server.HttpAddr, server.HttpPort))
	go configz.Heatload()
	go cache.List()
	go cache.Watch()
	return http.ListenAndServe(fmt.Sprintf("%v:%v", server.HttpAddr, server.HttpPort), nil)
}

func installApiGroup(router *mux.Router) {
	apiserver.InstallApi(router)
}
