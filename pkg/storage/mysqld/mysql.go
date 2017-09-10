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

package mysqld

import (
	"apiserver/pkg/configz"
	"apiserver/pkg/util/log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// --------------
// Engine

type DB struct {
	*gorm.DB
}

var (
	db *DB
)

func init() {
	tdb, err := gorm.Open("mysql", configz.GetString("mysql", "dsn", ""))
	if err != nil {
		log.Fatalf("init mysql connection err: %v", err)
	}
	db = &DB{DB: tdb}
}

func GetDB() *DB {
	return db
}
