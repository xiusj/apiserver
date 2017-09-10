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
	"encoding/json"
	"net/http"
	"strconv"

	author "apiserver/pkg/api/authorization"
	r "apiserver/pkg/router"

	"github.com/gorilla/mux"
)

func CreateRole(request *http.Request) (string, interface{}) {
	role, err := validateRole(request)
	if err != nil {
		return r.StatusBadRequest, err
	}

	if err = author.InsertRole(role); err != nil {
		return r.StatusInternalServerError, err
	}
	return r.StatusCreated, "ok"
}

func GetRoles(request *http.Request) (string, interface{}) {
	pageCnt, _ := strconv.Atoi(request.FormValue("pageCnt"))
	pageNum, _ := strconv.Atoi(request.FormValue("pageNum"))
	name := request.FormValue("name")
	roles, total, err := author.QueryRoles(name, pageCnt, pageNum)
	if err != nil {
		return r.StatusInternalServerError, err
	}
	return r.StatusOK, map[string]interface{}{"roles": roles, "total": total}
}

func DeleteRole(request *http.Request) (string, interface{}) {
	id, _ := strconv.ParseUint(mux.Vars(request)["id"], 10, 64)
	if err := author.DeleteRole(uint(id)); err != nil {
		return r.StatusInternalServerError, err
	}
	return r.StatusNoContent, "ok"
}

func validateRole(request *http.Request) (*author.Role, error) {
	role := &author.Role{}
	if err := json.NewDecoder(request.Body).Decode(role); err != nil {
		return nil, err
	}
	return role, nil
}
