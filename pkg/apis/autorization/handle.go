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

package autorization

import (
	"net/http"

	r "apiserver/pkg/router"

	"github.com/gorilla/mux"
)

func InstallApi(router *mux.Router) {
	//install user's api handle
	r.RegisterHttpHandler(router, "/users", "CREATE", CreateUser)
	r.RegisterHttpHandler(router, "/users", "GET", GetUsers)
	r.RegisterHttpHandler(router, "/users/{id}", "DELETE", DeleteUser)
	r.RegisterHttpHandler(router, "/users", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/users/{id}", "OPTIONS", Option)

	//install team's api handle
	r.RegisterHttpHandler(router, "/teams", "CREATE", CreateTeam)
	r.RegisterHttpHandler(router, "/teams", "GET", GetTeams)
	r.RegisterHttpHandler(router, "/teams/{id}", "DELETE", DeleteTeam)
	r.RegisterHttpHandler(router, "/teams", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/teams/{id}", "OPTIONS", Option)

	//install permission's api handle
	r.RegisterHttpHandler(router, "/permissions", "CREATE", CreatePermission)
	r.RegisterHttpHandler(router, "/permissions", "GET", GetPermissions)
	r.RegisterHttpHandler(router, "/permissions/{id}", "DELETE", DeletePermission)
	r.RegisterHttpHandler(router, "/permissions", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/permissions/{id}", "OPTIONS", Option)

	//install role's api handle
	r.RegisterHttpHandler(router, "/roles", "CREATE", CreateRole)
	r.RegisterHttpHandler(router, "/roles", "GET", GetRoles)
	r.RegisterHttpHandler(router, "/roles/{id}", "DELETE", DeleteRole)
	r.RegisterHttpHandler(router, "/roles", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/roles/{id}", "OPTIONS", Option)
}

func Option(request *http.Request) (string, interface{}) {
	return r.StatusOK, "ok"
}
