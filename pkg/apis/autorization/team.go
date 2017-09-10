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

func CreateTeam(request *http.Request) (string, interface{}) {
	team, err := validateTeam(request)
	if err != nil {
		return r.StatusBadRequest, err
	}

	if err = author.InsertTeam(team); err != nil {
		return r.StatusInternalServerError, err
	}
	return r.StatusCreated, "ok"
}

func GetTeams(request *http.Request) (string, interface{}) {
	pageCnt, _ := strconv.Atoi(request.FormValue("pageCnt"))
	pageNum, _ := strconv.Atoi(request.FormValue("pageNum"))
	name := request.FormValue("name")
	teams, total, err := author.QueryTeams(name, pageCnt, pageNum)
	if err != nil {
		return r.StatusInternalServerError, err
	}
	return r.StatusOK, map[string]interface{}{"teams": teams, "total": total}
}

func DeleteTeam(request *http.Request) (string, interface{}) {
	id, _ := strconv.ParseUint(mux.Vars(request)["id"], 10, 64)
	if err := author.DeleteTeam(uint(id)); err != nil {
		return r.StatusInternalServerError, err
	}
	return r.StatusNoContent, "ok"
}

func validateTeam(request *http.Request) (*author.Team, error) {
	team := &author.Team{}
	if err := json.NewDecoder(request.Body).Decode(team); err != nil {
		return nil, err
	}
	return team, nil
}
