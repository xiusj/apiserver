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
	"io"
	"net/http"
	"os"

	"apiserver/pkg/util/log"
)

/*func Upload(w http.ResponseWriter, r *http.Request, filePath string) {
	src, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.Copy(dst, src)
	w.Header().Add("Content-Type", "application/json")
	io.WriteString(w, `{"err"：0,"msg":"upload success","status":"success"}`)
}*/

func Upload(r *http.Request, fileDir string) error {
	src, head, err := r.FormFile("file")
	log.Debugf("err == %v", err)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(fileDir, os.ModePerm); err != nil {
		log.Errorf("create dir %v err: %v", fileDir, err)
		return err
	}
	defer src.Close()
	filePath := fileDir + string(os.PathSeparator) + head.Filename
	log.Debugf("dst == %v", filePath)
	dst, err := os.Create(filePath)
	log.Debugf("err == %v", err)
	if err != nil {
		log.Errorf("upload file %v err: %v", head.Filename, err)
		return err
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		log.Debugf("err == %v", err)
		log.Errorf("upload file %v err: %v", head.Filename, err)
		return err
	}
	return nil
}
