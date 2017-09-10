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

	"apiserver/pkg/apis/build"
	"apiserver/pkg/componentconfig"
	"apiserver/pkg/configz"
	"apiserver/pkg/util/log"

	"github.com/gorilla/mux"
	// "github.com/emicklei/go-restful"
)

type Buildserver struct {
	*componentconfig.BuildConfig
}

func NewBuildServer() *Buildserver {
	return &Buildserver{
		BuildConfig: &componentconfig.BuildConfig{
			HttpAddr: configz.GetString("build", "httpAddr", "0.0.0.0"),
			HttpPort: configz.MustInt("build", "httpPort", 9091),
			RpcAddr:  configz.GetString("build", "rpcAddr", "0.0.0.0"),
			RpcPort:  configz.MustInt("build", "rpcPort", 7071),
			Endpoint: configz.GetString("build", "endpoint", "http://127.0.0.1:2375"),
			Version:  configz.GetString("build", "version", "12.4"),
		},
	}
}

func Run(server *Buildserver) error {
	root := mux.NewRouter()
	api := root.PathPrefix("/api/v1").Subrouter()
	installApiGroup(api)
	http.Handle("/", root)
	log.Infof("starting buildserver and listen on : %v", fmt.Sprintf("%v:%v", server.HttpAddr, server.HttpPort))
	go configz.Heatload()
	return http.ListenAndServe(fmt.Sprintf("%v:%v", server.HttpAddr, server.HttpPort), nil)
}

func installApiGroup(router *mux.Router) {
	build.Register(router)
}
