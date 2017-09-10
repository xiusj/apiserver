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
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

//Tar tar some file to tar.gz file,it's same as: tar czf file1 file2 ....
//if fileExsit true will delete it and tar file ,if false will not do anything
func Tar(dstTar string, override bool, src ...string) (err error) {
	if !Exists(src...) {
		return errors.New(fmt.Sprintf("file %v doesn't exsit", strings.Join(src, ",")))
	}
	dstTar = path.Clean(dstTar)
	if Exists(dstTar) {
		if override {
			if err := os.Remove(dstTar); err != nil {
				return err
			}
		} else {
			return errors.New(fmt.Sprintf("the %v file exsited !", dstTar))
		}
	}
	filew, err := os.Create(dstTar)
	if err != nil {
		return err
	}
	defer filew.Close()
	gzipw := gzip.NewWriter(filew)
	defer gzipw.Close()
	tarw := tar.NewWriter(gzipw)
	defer tarw.Close()

	for _, s := range src {
		s = path.Clean(s)
		fi, err := os.Stat(s)
		if err != nil {
			return err
		}
		tarhead, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}
		if err = tarw.WriteHeader(tarhead); err != nil {
			return err
		}
		srcFile, err := os.Open(s)
		if err != nil {
			return err
		}
		defer srcFile.Close()
		if _, err = io.Copy(tarw, srcFile); err != nil {
			return err
		}
	}
	return nil
}

//Untar untar the tar file  to some dir
func UnTar(srcTar string, dstDir string) (err error) {
	srcTar = path.Clean(srcTar)
	if !FileExsit(srcTar) {
		return errors.New(fmt.Sprintf("file %v doesn't exsit", srcTar))
	}

	tarfile, err := os.Open(srcTar)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	gzipr, err := gzip.NewReader(tarfile)
	if err != nil {
		return err
	}
	defer gzipr.Close()
	tarr := tar.NewReader(gzipr)

	for head, err := tarr.Next(); err != io.EOF; head, err = tarr.Next() {
		fi := head.FileInfo()
		dstfile := dstDir + string(os.PathSeparator) + head.Name
		if head.Typeflag == tar.TypeDir {
			os.MkdirAll(dstfile, fi.Mode().Perm())
			os.Chmod(dstfile, fi.Mode().Perm())
		} else {
			os.MkdirAll(path.Dir(dstfile), os.ModePerm)
			df, err := os.Create(dstfile)
			if err != nil {
				return err
			}
			defer df.Close()
			if _, err = io.Copy(df, tarr); err != nil {
				return err
			}
		}
	}
	return nil
}
