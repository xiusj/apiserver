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

func QueryContainers(containerName string, pageCnt, pageNum int, serviceId uint) (list []*Container, total int) {

	if serviceId == 0 {
		if containerName != "" {
			db.Where("name like ? ", `%`+containerName+`%`).Offset(pageCnt * pageNum).Limit(pageCnt).Order("name desc").Find(&list)
			db.Model(new(Container)).Where("name like ?", containerName).Count(&total)
		} else {
			db.Offset(pageCnt * pageNum).Limit(pageCnt).Order("name desc").Find(&list)
			db.Model(new(Container)).Count(&total)
		}
	} else {
		if containerName != "" {
			db.Where("name like ? and service_id=?", `%`+containerName+`%`, serviceId).Offset(pageCnt * pageNum).Limit(pageCnt).Order("name desc").Find(&list)
			db.Model(new(Container)).Where("name like ? and service_id=?", `%`+containerName+`%`, serviceId).Count(&total)
		} else {
			db.Where("service_id=?", serviceId).Offset(pageCnt * pageNum).Limit(pageCnt).Order("name desc").Find(&list)
			db.Model(new(Container)).Where("service_id=?", serviceId).Count(&total)
		}
	}

	for _, container := range list {
		var (
			config      = &ContainerConfig{}
			base        = &BaseConfig{}
			configGroup = &ConfigGroup{}
			configmaps  = []*ConfigMap{}
			superConfig = &SuperConfig{}
			volumes     []*Volume
			envs        []*Env
			ports       []*Port
		)
		db.Find(config, ContainerConfig{ContainerId: container.ID})
		container.Config = config

		db.First(base, BaseConfig{ServiceConfigId: config.ID})
		db.Find(&volumes, Volume{BaseConfigId: base.ID})
		base.Volumes = volumes
		config.BaseConfig = base

		db.First(configGroup, ConfigGroup{ServiceConfigId: config.ID})
		config.ConfigGroup = configGroup

		db.Find(&configmaps, ConfigMap{ConfigGroupId: config.ID})
		configGroup.ConfigMaps = configmaps

		db.First(superConfig, SuperConfig{ServiceConfigId: config.ID})
		db.Find(&envs, Env{SuperConfigId: superConfig.ID})
		db.Find(&ports, Port{SuperConfigId: superConfig.ID})
		superConfig.Envs = envs
		superConfig.Ports = ports
		config.SuperConfig = superConfig
	}
	return
}

func QueryContainerById(id uint) *Container {
	container := &Container{}
	db.First(container, id)
	return container
}

func QueryContainerByName(name string) (*Container, bool) {
	container := &Container{}
	not := db.Where("name=?", name).First(container).RecordNotFound()
	return container, not
}

func UpdateContainer(container *Container) {
	db.Model(new(Container)).Update(container)
}

func DeleteContainer(container *Container) {
	db.Delete(container)
}

func ExistContainer(container *Container) bool {
	return db.First(container, "name=?", container.Name).RecordNotFound()
}

func InsertContainer(container *Container) {
	if db.Model(container).Where("name=?", container.Name).First(container).RecordNotFound() {
		db.Model(container).Save(container)
	}
}

func CountContainer() (interface{}, error) {
	var (
		ok    = 0
		stop  = 0
		fail  = 0
		build = 0
		err   error
	)
	err = db.Model(new(Container)).Where("status =?", 3).Count(&ok).Error
	err = db.Model(new(Container)).Not("status =?", 4).Count(&stop).Error
	err = db.Model(new(Container)).Not("status =?", 2).Count(&fail).Error
	err = db.Model(new(Container)).Not("status =?", 0).Count(&build).Error
	return map[string]int{"running": ok, "stop": stop, "fail": fail, "building": build}, err
}
