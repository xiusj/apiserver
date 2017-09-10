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

func QueryServices(serviceName string, pageCnt, pageNum int, appId uint) (list []*Service, total int) {

	if appId == 0 {
		if serviceName != "" {
			db.Where("name like ? ", `%`+serviceName+`%`).Offset(pageCnt * pageNum).Limit(pageCnt).Order("name desc").Find(&list)
			db.Model(new(Service)).Where("name like ?", serviceName).Count(&total)
		} else {
			db.Offset(pageCnt * pageNum).Limit(pageCnt).Order("name desc").Find(&list)
			db.Model(new(Service)).Count(&total)
		}
	} else {
		if serviceName != "" {
			db.Where("name like ? and app_id=?", `%`+serviceName+`%`, appId).Offset(pageCnt * pageNum).Limit(pageCnt).Order("name desc").Find(&list)
			db.Model(new(Service)).Where("name like ? and app_id=?", `%`+serviceName+`%`, appId).Count(&total)
		} else {
			db.Where("app_id=?", appId).Offset(pageCnt * pageNum).Limit(pageCnt).Order("name desc").Find(&list)
			db.Model(new(Service)).Where("app_id=?", appId).Count(&total)
		}
	}

	for _, svc := range list {
		if svc.ID != 0 {
			var (
				containerList []*Container
				config        = &ContainerConfig{}
				serviceConfig = &ServiceConfig{}
				base          = &BaseConfig{}
				configGroup   = &ConfigGroup{}
				configmaps    = []*ConfigMap{}
				superConfig   = &SuperConfig{}
				volumes       []*Volume
				envs          []*Env
				ports         []*Port
			)
			db.Find(&containerList, Container{ServiceId: svc.ID})
			svc.Items = containerList
			for _, container := range containerList {
				db.Find(config, ContainerConfig{ContainerId: container.ID})
				// container.Config = config

				db.First(base, BaseConfig{ServiceConfigId: config.ID})
				db.Find(&volumes, Volume{BaseConfigId: base.ID})
				base.Volumes = volumes
				// config.BaseConfig = base

				db.First(configGroup, ConfigGroup{ServiceConfigId: config.ID})

				db.Find(&configmaps, ConfigMap{ConfigGroupId: configGroup.ID})
				configGroup.ConfigMaps = configmaps

				db.First(superConfig, SuperConfig{ServiceConfigId: config.ID})
				db.Find(&envs, Env{SuperConfigId: superConfig.ID})
				db.Find(&ports, Port{SuperConfigId: superConfig.ID})
				superConfig.Envs = envs
				superConfig.Ports = ports
				// config.SuperConfig = superConfig
			}
			db.First(serviceConfig, ServiceConfig{ServiceId: svc.ID})
			serviceConfig.BaseConfig = base
			serviceConfig.ConfigGroup = configGroup
			serviceConfig.SuperConfig = superConfig
			svc.Config = serviceConfig
		}
	}
	return
}

func InsertService(svc *Service) {
	svcConfig := svc.Config
	if len(svc.Items) != 0 {
		svc.Items[0].Config = &ContainerConfig{
			BaseConfig:  svcConfig.BaseConfig,
			SuperConfig: svcConfig.SuperConfig,
		}
	}

	if svcConfig.ConfigGroup != nil {
		for _, c := range svcConfig.ConfigGroup.ConfigMaps {
			UpdateConfigMap(c)
		}
	}

	configGroupId := svcConfig.ConfigGroup.ID
	svcConfig.ConfigGroup = nil

	if db.Model(svc).Where("name=?", svc.Name).First(svc).RecordNotFound() {
		db.Model(svc).Create(svc)
	}

	db.Model(new(ConfigGroup)).Set("gorm:save_associations", false).Update(&ConfigGroup{ServiceConfigId: svc.Config.ID, ServiceName: svc.Name, ID: configGroupId})
}

func DeleteService(svc *Service) {
	db.Delete(svc)
	svcCfg := svc.Config
	db.Delete(svcCfg, "service_id=?", svc.ID)

	svcCfgBase := svcCfg.BaseConfig
	db.Delete(svcCfgBase, "service_config_id=?", svcCfg.ID)

	for _, volume := range svcCfgBase.Volumes {
		db.Delete(volume, "base_config_id=?", svcCfgBase.ID)
	}

	svcSuper := svcCfg.SuperConfig
	db.Delete(svcSuper, "service_config_id=?", svcCfg.ID)

	for _, env := range svcSuper.Envs {
		db.Delete(env, "super_config_id=?", svcSuper.ID)
	}
	for _, port := range svcSuper.Ports {
		db.Delete(port, "super_config_id=?", svcSuper.ID)
	}

	for _, c := range svc.Items {
		db.Delete(c)
		if c.Config != nil {
			db.Delete(c.Config, "container_id=?", c.ID)
		}
	}

	if svc.Config.ConfigGroup != nil {
		for _, c := range svc.Config.ConfigGroup.ConfigMaps {
			c.ContainerPath = ""
		}
		db.Model(new(ConfigGroup)).Update(svc.Config.ConfigGroup)
		db.Exec("update config_group set service_config_id = ?,service_name=?", 0, "")
	}
}

func UpdateService(svc *Service) {
	db.Model(svc).Set("gorm:save_associations", false).Update(svc)
}

func QueryServiceById(id uint) *Service {
	svc := &Service{}
	var (
		containerList []*Container
		serviceConfig = &ServiceConfig{}
		base          = &BaseConfig{}
		configGroup   = &ConfigGroup{}
		configmaps    = []*ConfigMap{}
		superConfig   = &SuperConfig{}
		volumes       []*Volume
		envs          []*Env
		ports         []*Port
	)
	if db.First(svc, id).Error == nil && svc.ID != 0 {
		db.Find(&containerList, Container{ServiceId: svc.ID})
		svc.Items = containerList

		db.Find(serviceConfig, ServiceConfig{ServiceId: svc.ID})
		svc.Config = serviceConfig

		db.First(base, BaseConfig{ServiceConfigId: svc.ID})
		db.Find(&volumes, Volume{BaseConfigId: base.ID})
		base.Volumes = volumes
		serviceConfig.BaseConfig = base

		db.First(configGroup, ConfigGroup{ServiceConfigId: serviceConfig.ID})
		serviceConfig.ConfigGroup = configGroup

		db.Find(&configmaps, ConfigMap{ConfigGroupId: configGroup.ID})
		configGroup.ConfigMaps = configmaps

		db.First(superConfig, SuperConfig{ServiceConfigId: serviceConfig.ID})
		db.Find(&envs, Env{SuperConfigId: superConfig.ID})
		db.Find(&ports, Port{SuperConfigId: superConfig.ID})
		superConfig.Envs = envs
		superConfig.Ports = ports
		serviceConfig.SuperConfig = superConfig
	}

	return svc
}

func ExistService(svc *Service) bool {
	return !db.First(svc).RecordNotFound()
}

func QueryServicesByAppId(id uint) []*Service {
	svcs := []*Service{}
	db.Find(&svcs, "app_id=?", id)
	return svcs
}

func UpdateServiceOnly(svc *Service) {
	db.Model(svc).Set("gorm:save_associations", false).Update(svc)
}

func CountService() (interface{}, error) {
	var (
		ok    = 0
		stop  = 0
		fail  = 0
		build = 0
		err   error
	)
	err = db.Model(new(Service)).Where("status =?", 3).Count(&ok).Error
	err = db.Model(new(Service)).Not("status =?", 4).Count(&stop).Error
	err = db.Model(new(Service)).Not("status =?", 2).Count(&fail).Error
	err = db.Model(new(Service)).Not("status =?", 0).Count(&build).Error
	return map[string]int{"running": ok, "stop": stop, "fail": fail, "building": build}, err
}
