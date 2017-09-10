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

func InsertRole(role *Role) error {
	return db.Create(role).Error
}

func QueryRoles(name string, pageCnt, pageNum int) (list []*Role, total int, err error) {
	if name != "" {
		err = db.Where("name like ? ", `%`+name+`%`).Offset(pageCnt * pageNum).Limit(pageCnt).Order("name desc").Find(&list).Error
		db.Model(new(Role)).Where("name like ?", name).Count(&total)
	} else {
		err = db.Offset(pageCnt * pageNum).Limit(pageCnt).Order("name desc").Find(&list).Error
		db.Model(new(Role)).Count(&total)
	}
	return
}

func DeleteRole(id uint) error {
	return db.Model(new(Role)).Delete(new(Role), id).Error
}
