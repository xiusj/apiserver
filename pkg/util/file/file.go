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

package file

import (
	"os"
)

//Exsists assert the dir or file exsit or not
func Exists(name ...string) (exsit bool) {
	for _, n := range name {
		_, err := os.Stat(n)
		if err == nil || os.IsExist(err) {
			exsit = true
		}
	}
	return
}

//FileExsit assert file exsit or not
func FileExsit(file ...string) (exsit bool) {
	for _, f := range file {
		fi, err := os.Stat(f)
		if (err == nil && os.IsExist(err)) || !fi.IsDir() {
			exsit = true
		}
	}
	return
}

//DirExsit assert dir exsit or not
func DirExsit(dir ...string) (exsit bool) {
	for _, d := range dir {
		fi, err := os.Stat(d)
		if (err == nil && os.IsExist(err)) || fi.IsDir() {
			exsit = true
		}
	}
	return
}
