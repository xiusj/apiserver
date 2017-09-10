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

import ()

// func InsertConfigs(c *ConfigMap) {
// 	db.Model(new(ConfigMap)).Create(c)
// }

// func DeleteConfigMap() {

// }

// func QueryConfigs(namespace, appName string, pageCnt, pageNum int) (list []*Config, total int) {
// 	return
// }

func UpdateConfig(c *ConfigGroup) {
	db.Model(c).Update(c)
}

func UpdateConfigMap(c *ConfigMap) {
	db.Model(c).Update(c)
}

func DeleteConfig(id uint) {
	db.Model(new(ConfigGroup)).Delete(new(ConfigGroup), id)
	db.Model(new(ConfigMap)).Delete(new(ConfigMap), "config_group_id=? ", id)
}

func DeleteConfigItem(id uint) {
	db.Model(new(ConfigMap)).Delete(new(ConfigMap), id)
}

func InsertConfig(c *ConfigGroup) {
	db.Model(c).Create(c)
}

func InsertConfigItem(c *ConfigMap) {
	db.Model(c).Create(c)
}

func QueryConfigById(id uint) *ConfigGroup {
	cfg := &ConfigGroup{}
	db.Model(new(ConfigMap)).First(cfg, id)

	var cfgmaps []*ConfigMap
	db.Model(new(ConfigMap)).Find(&cfgmaps, "config_group_id=?", cfg.ID)
	cfg.ConfigMaps = cfgmaps
	return cfg
}

func QueryConfigs(configName string, pageCnt, pageNum int) (list []*ConfigGroup, total int) {
	if configName != "" {
		db.Where("name like ? ", `%`+configName+`%`).Offset(pageCnt * pageNum).Limit(pageCnt).Order("name desc").Find(&list)
		db.Model(new(ConfigGroup)).Where("name like ?", configName).Count(&total)
	} else {
		db.Offset(pageCnt * pageNum).Limit(pageCnt).Order("name desc").Find(&list)
		db.Model(new(ConfigGroup)).Count(&total)
	}

	for _, cfg := range list {
		var cfgmaps []*ConfigMap
		db.Model(new(ConfigMap)).Find(&cfgmaps, "config_group_id=?", cfg.ID)
		cfg.ConfigMaps = cfgmaps
	}

	return
}
