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

package configz

import (
	"os"

	"apiserver/pkg/util/log"

	"github.com/Unknwon/goconfig"
	"github.com/howeyc/fsnotify"
)

var (
	cfg *goconfig.ConfigFile
	err error
)

//init load the application's conifg file
func init() {
	config := os.Getenv("CONFIG_PATH")
	if config == "" {
		panic("you should define CONFIG_PATH environment.")
	}
	cfg, err = goconfig.LoadConfigFile(config)
	if err != nil {
		panic(err)
	}
}

func GetString(section, key, defaults string) string {
	return cfg.MustValue(section, key, defaults)
}

func GetStringArray(section, key, delim string) []string {
	return cfg.MustValueArray(section, key, delim)
}

func MustBool(section, key string, defaultVal bool) bool {
	return cfg.MustBool(section, key, defaultVal)
}

func MustFloat64(section, key string, defaultVal float64) float64 {
	return cfg.MustFloat64(section, key, defaultVal)
}

func MustInt(section, key string, defaultVal int) int {
	return cfg.MustInt(section, key, defaultVal)
}

func MustInt64(section, key string, defaultVal int64) int64 {
	return cfg.MustInt64(section, key, defaultVal)
}

//watcher notify the config file, when the file was changed, reload the file to memory
func Heatload() {
	config := os.Getenv("CONFIG_PATH")
	wacther, err := fsnotify.NewWatcher()
	if err != nil {
		log.Errorf("create the file watcher err: %v", err)
	}
	defer func() {
		if err = wacther.Close(); err != nil {
			log.Fatalf("close the file wather err:%v", err)
		}
	}()

	wacther.Watch(config)
	done := make(chan bool)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("panic is happend: %v", err)
			}
		}()
		for {
			select {
			case event := <-wacther.Event:
				if event.IsCreate() || event.IsModify() || event.IsAttrib() {
					cfg, err = goconfig.LoadConfigFile(config)
					if err != nil {
						panic(err)
					}
				}
			case err := <-wacther.Error:
				log.Errorf("the file watcher err: %v", err)
			}
		}
	}()
	<-done
}
