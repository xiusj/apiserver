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

package storageclass

import (
	"apiserver/pkg/configz"
	"apiserver/pkg/resource"

	"k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewStorageClass(useraName string) *v1.StorageClass {
	return &v1.StorageClass{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StorageClass",
			APIVersion: "storage.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      useraName,
			Namespace: useraName,
		},
		Provisioner: resource.Provisioner_ceph_rbd,
		Parameters: map[string]string{
			"monitors":             configz.GetString("ceph", "monitors", ""),
			"adminId":              configz.GetString("ceph", "adminId", ""),
			"adminSecretName":      configz.GetString("ceph", "adminSecretName", ""),
			"adminSecretNamespace": configz.GetString("ceph", "adminSecretNamespace", ""),
			"pool":                 configz.GetString("ceph", "pool", ""),
			"userId":               configz.GetString("ceph", "userId", ""),
			"userSecretName":       configz.GetString("ceph", "userSecretName", ""),
		},
	}
}
