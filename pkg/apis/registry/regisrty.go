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

package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	regModel "apiserver/pkg/api/registry"
	"apiserver/pkg/configz"
	r "apiserver/pkg/router"
	"apiserver/pkg/util/log"
	regUtil "apiserver/pkg/util/registry"

	"github.com/gorilla/mux"
)

func Register(router *mux.Router) {
	r.RegisterHttpHandler(router, "/{namespace}/images", "GET", GetImages)
	r.RegisterHttpHandler(router, "/{namespace}/{name}/images/{reference}", "Delete", DeleteImage)
	r.RegisterHttpHandler(router, "/{namespace}/{name}/images/{reference}", "OPTIONS", Option)
	r.RegisterHttpHandler(router, "/{namespace}/images", "OPTIONS", Option)

	//install apiserver's system api,include health check and get apiserver's version api
	r.RegisterHttpHandler(router, "/registry/health", "GET", Health)
	r.RegisterHttpHandler(router, "/registry/version", "GET", GetRegistryInfo)
	// r.RegisterHttpHandler(router, "/images", "DELETE", DeleteImage)
}

var (
	registry *regUtil.Registry
	err      error
	store    *storage
)

type storage struct {
	lock  sync.RWMutex
	cache map[string]interface{}
}

func init() {
	tranport := regUtil.GetHttpTransport(false)
	endpoint := configz.GetString("registry", "endpoint", "http://0.0.0.0:5000")
	client := &http.Client{Transport: tranport}
	if registry, err = regUtil.NewRegistry(endpoint, client); err != nil {
		log.Fatalf("init resgistry err: %v", err)
	}
	store = &storage{cache: map[string]interface{}{}}
	go task()
}

func Option(request *http.Request) (string, interface{}) {
	return r.StatusOK, nil
}

func GetImages(req *http.Request) (string, interface{}) {
	namespace := mux.Vars(req)["namespace"]
	name := req.FormValue("name")
	pageCnt := req.FormValue("pageCnt")
	pageNum := req.FormValue("pageNum")
	cnt, _ := strconv.Atoi(pageCnt)
	num, _ := strconv.Atoi(pageNum)
	set, total := new(regModel.Manifest).QuerySet(map[string]interface{}{"namespace": namespace, "name": name, "pageCnt": cnt, "pageNum": num})
	imageSet := []*regModel.Image{}
	if tags, exsit := store.cache[name]; exsit {
		image := &regModel.Image{}
		for _, m := range set {
			image.Name = name
			image.TagLen = len(tags.([]string))
			image.Tags = tags.([]string)
			image.Fest = append(image.Fest, m)
		}
		if image.Name != "" {
			imageSet = append(imageSet, image)
		}
	} else {
		for k, v := range store.cache {
			tgs := []string{}
			image := &regModel.Image{}
			for _, m := range set {
				if k == m.Name {
					image.Name = k
					image.TagLen = len(v.([]string))
					tgs = append(tgs, m.Tag)
					image.Fest = append(image.Fest, m)
				}
			}
			image.Tags = tgs
			if image.Name != "" {
				imageSet = append(imageSet, image)
			}
		}
	}

	return r.StatusOK, map[string]interface{}{"images": imageSet, "total": total}
}

func DeleteImage(req *http.Request) (string, interface{}) {
	reference := req.FormValue("reference")
	catalog := req.FormValue("name")
	if err := registry.DelManifest(catalog, reference); err != nil {
		return r.StatusInternalServerError, err
	}
	m := &regModel.Manifest{Id: reference}
	err = m.Delete()
	if err != nil {
		return r.StatusInternalServerError, err
	}
	return r.StatusOK, nil
}

func task() {
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ticker.C:
			go SyncImage()
		}
	}
}

func SyncImage() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("panic is occur :%v", err)
		}
	}()
	catalogs, err := registry.GetCatalog()
	if err != nil {
		log.Errorf("get catalogs err:%v", err)
		return
	}
	for _, catalog := range catalogs {
		store.lock.Lock()
		tags, err := registry.GetTags(catalog)
		if err != nil {
			log.Errorf("get the tags catalog named %v  err:%v", catalog, err)
			continue
		}
		store.cache[catalog] = tags
		store.lock.Unlock()
		for _, tag := range tags {
			manifest, err := registry.GetManifest(catalog, tag)
			if err != nil {
				log.Errorf("get the mainifest of  named %v and reference named %v's err:%v", catalog, tag, err)
				continue
			}
			ms := &regModel.Manifest{}
			if err = json.Unmarshal([]byte(manifest.History[0].V1Compatibility), ms); err != nil {
				log.Errorf("unmarshal manifest err:%v", err)
				continue
			}
			ms.Name = catalog
			ms.Tag = tag
			ms.Pull = fmt.Sprintf("docker pull %s/%s:%s", configz.GetString("registry", "endpoint", "http://0.0.0.0:5000"), catalog, tag)
			if exsit := ms.Exsit(); !exsit {
				ms.Insert()
			}
		}
	}
}
