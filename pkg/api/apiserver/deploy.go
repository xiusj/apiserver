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
	"apiserver/pkg/util/log"
)

func InsertDeploy(deploy *Deploy) error {
	return db.Create(deploy).Error
}

func InsertProjectConfig(config *ProjectConfig) error {
	return db.Create(config).Error
}

func QueryDeployByImage(image, tag string) *Deploy {
	var item *DeployItem
	if err := db.Model(new(DeployItem)).Where("docker_repo_url=? and tag=?", image, tag).First(&item).Error; err != nil {
		log.Debug(err)
		return nil
	}

	var deploy *Deploy
	if item != nil {
		if err := db.First(&deploy, item.DeployId).Error; err != nil {
			log.Debug(err)
			return nil
		}
		deploy.Items = []*DeployItem{item}
	}
	return deploy
}

func QueryProjectConfigByID(id uint) *ProjectConfig {
	projectConfig := &ProjectConfig{}
	db.Model(new(ProjectConfig)).Where("project_id=?", id).First(projectConfig)
	return projectConfig
}

func QueryNoDeployItem() ([]*DeployItem, int) {
	var Items []*DeployItem
	var total int
	db.Model(new(DeployItem)).Where("status=?", 0).Find(&Items)
	db.Model(new(DeployItem)).Where("status=?", 0).Count(&total)
	return Items, total
}
