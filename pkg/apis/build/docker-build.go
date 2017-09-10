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

package build

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"apiserver/pkg/api/build"
	"apiserver/pkg/client"
	r "apiserver/pkg/router"
	"apiserver/pkg/util/file"
	"apiserver/pkg/util/log"

	"github.com/docker/docker/api/types"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

const (
	DEFAULT_REGISTRY = `hub.mini-paas.com`
	/*	TARBALL_ROOT_DIR    = `/tmp`
		BUILD_IMAGE_TAR_DIR = `/tmp/build`*/
	//incorde to test,my env is windows ,so l use this
	TARBALL_ROOT_DIR    = `F:\workspace\src\huangjia`
	BUILD_IMAGE_TAR_DIR = `F:\workspace\src\huangjia`
)

func Register(rout *mux.Router) {
	r.RegisterHttpHandler(rout, "/builds", "POST", OnlineBuild)
	r.RegisterHttpHandler(rout, "/builds", "PUT", OfflineBuild)
	r.RegisterHttpHandler(rout, "/uploads", "POST", Upload)
}

//OnlineBuild build application online
func OnlineBuild(request *http.Request) (string, interface{}) {
	decoder := json.NewDecoder(request.Body)
	builder := &build.Build{}
	err := decoder.Decode(builder)
	if err != nil {
		log.Errorf("decode the request body err:%v", err)
		return r.StatusBadRequest, "json format error"
	}
	//TODO
	return r.StatusCreated, nil
}

//OfflineBuild build application offline
////build image's step:
//1. find the upload tarball file of user uploaded in TARBALL_ROOT_DIR/userid/ dir
//2. find the Dockerfile file of the user upload in TARBALL_ROOT_DIR/userid/ dir
//3. if the Dockerfile exsit,we will use the tarball and the Dockerfile to create a tar file's stream in order to build image.
//4. if the Dokcerfile didn't exsit,will be use the Dockerfile template to create Dockerfile by the project's type. for exampler:if
//the project's language is golang, we will use the golang's Dockerfile template,and then we will use the tarball and the Dockerfile
//to create a tar file's stream in order to build image.
func OfflineBuild(request *http.Request) (string, interface{}) {
	decoder := json.NewDecoder(request.Body)
	builder := &build.Build{}
	err := decoder.Decode(builder)
	if err != nil {
		log.Errorf("decode the request body err:%v", err)
		return r.StatusBadRequest, "json format error"
	}

	dockerfile := fmt.Sprintf("%s/%s/%s", TARBALL_ROOT_DIR, builder.UserId, "Dockerfile")
	if !file.FileExsit(dockerfile) {
		//TODO generate the Dockerfile by Dockerfile template
	}
	projects := fmt.Sprintf("%s/%s/%s", TARBALL_ROOT_DIR, builder.UserId, builder.Tarball)
	imgBuildTar := fmt.Sprintf("%s/%s/%s", BUILD_IMAGE_TAR_DIR, builder.UserId, builder.Tarball)
	if err = file.Tar(imgBuildTar, true, dockerfile, projects); err != nil {
		return r.StatusInternalServerError, err
	}
	buildContext, err := os.Open(imgBuildTar)
	if err != nil {
		return r.StatusInternalServerError, err
	}
	defer buildContext.Close()
	image_repo := fmt.Sprintf("%s/%s:%s", DEFAULT_REGISTRY, builder.AppName, builder.Version)
	options := types.ImageBuildOptions{
		Tags:       []string{image_repo},
		Dockerfile: "Dockerfile",
	}

	buildResponse, err := client.DockerClient.ImageBuild(context.Background(), buildContext, options)
	if err != nil {
		log.Errorf("build image err: %v", err)
		return r.StatusInternalServerError, err.Error()
	}
	res, err := ioutil.ReadAll(buildResponse.Body)
	if err != nil {
		log.Errorf("read the build image response err: %v", err)
		return r.StatusInternalServerError, err.Error()
	}

	builder.Image = image_repo
	builder.Status = build.BUILD_SUCCESS
	builder.BuildLog = string(res)
	if err = builder.Insert(); err != nil {
		log.Errorf("insert the build to db err: %v", err)
	}

	pushRes, err := pushImage(image_repo)
	if err != nil {
		return r.StatusInternalServerError, "build image successed,but push image to registry err :" + err.Error()
	}
	log.Debugf("push result ==%v", pushRes)
	return r.StatusCreated, string(res)
}

//BuildProject when the online build image , build the github or gitlab's project resource code
//before the use call the online build api
func BuildProject(request *http.Request) (string, interface{}) {
	//the build step:
	//1. get the resouce code's repo and branch
	//2. select the build env base image acording to the projects language
	//3. build the project and output the tar or binary file to a appoint dir
	//4. if the project include Dockerfile,and then output the Dockerfile together
	//5. if the project doesn't include Dockerfile,and the generate the Dockerfile by Dockerfile templaet
	//TODO
	return r.StatusOK, "build the project resource success"
}

func pushImage(image string) (string, error) {
	res, err := client.DockerClient.ImagePush(context.Background(), image, types.ImagePushOptions{})
	if err != nil {
		return "", err
	}
	pushResult, err := ioutil.ReadAll(res)
	if err != nil {
		log.Errorf("read the build image response err: %v", err)
		return r.StatusInternalServerError, err
	}
	return string(pushResult), err
}

func Upload(request *http.Request) (string, interface{}) {
	// filePath := fmt.Sprintf("%s/%s/%s", TARBALL_ROOT_DIR, builder.UserId, "Dockerfile")
	//when the autorization is finished , get UserId from session,but now write is U001
	fileDir := fmt.Sprintf("%s/%s", TARBALL_ROOT_DIR, "U001")
	if err := file.Upload(request, fileDir); err != nil {
		return r.StatusInternalServerError, err.Error()
	}
	return r.StatusCreated, "upload file successed"
}
