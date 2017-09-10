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
	"errors"
	"time"

	"apiserver/pkg/storage/mysqld"
	"apiserver/pkg/util/jsonx"
	"apiserver/pkg/util/log"
)

const (
	BUILDING      = 0
	BUILD_SUCCESS = 1
	BUILD_FAILED  = 2
)

type Build struct {
	Id         int       `json:"id" xorm:"pk not null autoincr"`
	AppName    string    `json:"app_name" xorm:"not null varchar(255)"`
	Version    string    `json:"version" xorm:"not null varchar(255)"`
	Remark     string    `json:"remark" xorm:"not null varchar(255)"`
	BaseImage  string    `json:"baseImage" xorm:"not null varchar(255)"`
	Image      string    `json:"image" xorm:"not null varchar(255)"`
	Tarball    string    `json:"tarball" xorm:"not null varchar(255)"`
	Registry   string    `json:"registry" xorm:"not null varchar(255)"`
	Repositroy string    `json:"repository" xorm:"not null varchar(255)"`
	Branch     string    `json:"branch" xorm:"not null varchar(255)"`
	Status     int       `json:"status" xorm:"not null int(1)"` //0 构建中 1 构建成功 2 构建失败
	BuildLog   string    `json:"buildLog" xorm:"not null varchar(4096)"`
	Create_At  time.Time `json:"create_at" xorm:"created not null"`
	UserId     string    `json:"userId" xorm:"not null varchar(255)"`
	Language   string    `json:"language" xorm:"not null varchar(255)"`
}

var (
	engine = mysqld.GetEngine()
)

func init() {
	if err := engine.Sync(new(Build)); err != nil {
		log.Fatalf("Sync fail :%s", err.Error())
	}
}

func (this *Build) String() string {
	buildStr := jsonx.ToJson(this)
	return buildStr
}

func (this *Build) Insert() error {
	_, err := engine.Insert(this)
	if err != nil {
		return err
	}
	return nil
}

func (this *Build) Delete() error {
	_, err := engine.Id(this.Id).Delete(this)
	if err != nil {
		return err
	}
	return nil
}

func (this *Build) Update() error {
	_, err := engine.Id(this.Id).Update(this)
	if err != nil {
		return err
	}
	return nil
}

func (this *Build) QueryOne() (*Build, error) {
	has, err := engine.Id(this.Id).Get(this)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("current app not exsit")
	}
	return this, nil
}

func (this *Build) QuerySet() ([]*Build, error) {
	buildSet := []*Build{}
	err := engine.Where("1 and 1 order by create_at desc").Find(&buildSet)
	if err != nil {
		return nil, err
	}
	return buildSet, nil
}
