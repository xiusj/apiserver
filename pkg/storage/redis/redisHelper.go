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

package util

import (
	"errors"
	"time"

	"apiserver/pkg/configz"

	"gopkg.in/redis.v5"
)

var (
	client *redis.Client
)

func init() {
	addr := configz.GetString("redis", "address", "0.0.0.0:6379")
	password := configz.GetString("redis", "password", "")
	db := configz.MustInt("redis", "db", 0)
	poolSize := configz.MustInt("redis", "poolSize", 10)
	if configz.MustBool("redis", "requiered_password", false) {
		client = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
			PoolSize: poolSize,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:     addr,
			DB:       db,
			PoolSize: poolSize,
		})
	}
}

func Set(key string, val interface{}, expiredTime int) (string, error) {
	result, err := client.Set(key, val, time.Minute*30).Result()
	if err != nil {
		return "", err
	}
	return result, nil
}

func Get(key string) (string, error) {
	result, err := client.Get(key).Result()
	if err != nil {
		return "", err
	}
	if result == "" {
		return "", errors.New("key's value is not exsit")
	}
	return result, nil
}

func MutilSet() {
	// this.Client.Pipelined(fn).
}
