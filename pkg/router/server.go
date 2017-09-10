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

package router

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"apiserver/pkg/util/log"
	"github.com/gorilla/mux"
)

// https://golang.org/src/net/http/status.go
var (
	// 200
	StatusOK = strconv.Itoa(http.StatusOK)

	// 201
	StatusCreated = strconv.Itoa(http.StatusCreated)

	// 204
	StatusNoContent = strconv.Itoa(http.StatusNoContent)

	// 400
	StatusBadRequest = strconv.Itoa(http.StatusBadRequest)

	// 402
	StatusPaymentRequired = strconv.Itoa(http.StatusPaymentRequired)

	// 403
	StatusForbidden = strconv.Itoa(http.StatusForbidden)

	// 404
	StatusNotFound = strconv.Itoa(http.StatusNotFound)

	// 409
	StatusConflict = strconv.Itoa(http.StatusConflict)

	// 500
	StatusInternalServerError = strconv.Itoa(http.StatusInternalServerError)
)

const (
	HTTP_GET    = "GET"
	HTTP_POST   = "POST"
	HTTP_PUT    = "PUT"
	HTTP_DELETE = "DELETE"
)

const (
	OK               = "OK"
	JSON_EMPTY_ARRAY = "[]"
	JSON_EMPTY_OBJ   = "{}"
)

type HttpHandler func(*http.Request) (string, interface{})

func RegisterHttpHandler(router *mux.Router, path, method string, handler HttpHandler) {
	h := func(w http.ResponseWriter, r *http.Request) {
		// parseForm
		if err := r.ParseForm(); err != nil {
			log.Warning(err)
		}

		// dump
		bytes, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Warning(err)
		} else {
			log.Debug(string(bytes))
		}

		dump := dumpHttpRequest(r)
		log.Debug(fmt.Sprintf("incoming %s %s request from %v", r.Method, r.URL.RequestURI(), r.Host))

		t := time.Now()
		status, body := handler(r)
		writeHttpResp(w, dump, r.Host, status, body, t)
	}
	router.HandleFunc(path, h).Methods(method)
}

func dumpHttpRequest(r *http.Request) string {
	if r.Method == "GET" {
		return fmt.Sprintf("%s %s", r.Method, r.URL.RequestURI())
	}

	if r.Method == "POST" {
		return fmt.Sprintf("%s %s", r.Method, r.URL.RequestURI())
	}

	return fmt.Sprintf("%s %s %s", r.Method, r.URL.RequestURI(), r.Form)
}

// --------------------------------
// response

const httpJsonRespFmt = `{
  "api": "1.0",
  "status": "%v",
  "err": %v,
  "msg": %v
}
`

type httpRes struct {
	Apiversion string      `json:"apiversion"`
	Status     string      `json:"status"`
	Err        string      `json:"err,omitempty"`
	Data       interface{} `json:"data,omitempty"`
}

func writeHttpResp(w http.ResponseWriter, dump, host string, status string, body interface{}, t time.Time) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, X-Registry-Auth")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,OPTIONS,PUT,DELETE")
	sub := time.Now().Sub(t)
	/*	sub := time.Now().Sub(t)
		// empty array
		if body == JSON_EMPTY_ARRAY {
			log.Info(dump, status, sub)
			fmt.Fprintf(w, httpJsonRespFmt, status, `""`, body)
			return
		}

		if body == JSON_EMPTY_OBJ {
			log.Info(dump, status, sub)
			fmt.Fprintf(w, httpJsonRespFmt, status, `""`, body)
			return
		}

		errStr, data := "", JSON_EMPTY_OBJ
		res, err := json.MarshalIndent(body, " ", "    ")
		if err != nil {
			errStr = `"` + err.Error() + `"`
			log.Debug(dump, status, errStr, data, sub)
			fmt.Fprintf(w, httpJsonRespFmt, status, errStr, data)
			return
		}

		// error
		if status != StatusOK && status != StatusCreated && status != StatusNoContent {
			errStr = string(res)
			log.Debug(dump, status, errStr, data, sub)
			fmt.Fprintf(w, httpJsonRespFmt, status, errStr, data)
			return
		}

		errStr = `"` + OK + `"`
		data = string(res)

		log.Debug(dump, status, sub)
		fmt.Fprintf(w, httpJsonRespFmt, status, errStr, data)*/
	res := &httpRes{Apiversion: "v1", Status: status}
	if err, ok := body.(error); ok {
		res.Err = err.Error()
	} else {
		res.Data = body
	}

	resdate, _ := json.MarshalIndent(res, " ", "    ")
	log.Debug("outcoming response to ", host, dump, status, sub)
	io.WriteString(w, string(resdate))
}
