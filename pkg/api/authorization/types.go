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
	"time"
)

type Team struct {
	ID       uint      `json:"id,omitempty"`
	Name     string    `json:"name,omitempty"`
	CreateAt time.Time `json:"createAt,omitempty"`
	Users    []*User   `json:"users,omitempty"`
}

type User struct {
	ID       uint      `json:"id,omitempty"`
	Name     string    `json:"name,omitempty"`
	Password string    `json:"paasWord,omitempty"`
	Mail     string    `json:"mail,omitempty"`
	Phone    string    `json:"phone,omitempty"`
	Count    int       `json:"count,omitempty"`
	CreateAt time.Time `json:"createAt,omitempty"`
	TeamId   uint      `json:"teamId,omitempty"`
	RoleId   uint      `json:"roleId,omitempty"`
}

type Role struct {
	ID          uint          `json:"id,omitempty"`
	Name        string        `json:"name,omitempty"`
	CreateAt    time.Time     `json:"createAt,omitempty"`
	Permissions []*Permission `json:"permissions,omitempty"`
}

type Permission struct {
	ID       uint      `json:"id,omitempty"`
	Name     string    `json:"name,omitempty"`
	Url      string    `json:"url,omitempty"`
	CreateAt time.Time `json:"createAt,omitempty"`
	RoleId   uint      `json:"roleId,omitempty"`
}
