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
	"apiserver/pkg/storage/mysqld"
)

var (
	db = mysqld.GetDB()
)

func init() {
	db.SingularTable(true)
	db.CreateTable(
		new(App),
		new(Service),
		new(Container),
		new(Port),
		new(Env),
		new(SuperConfig),
		new(ConfigMap),
		new(Volume),
		new(BaseConfig),
		new(ServiceConfig),
		new(ContainerConfig),
		new(ConfigGroup),
		new(Deploy),
		new(DeployItem),
		new(ProjectConfig),
		new(Result),
		new(ResultItem),
	)
}
