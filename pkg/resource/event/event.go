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

package event

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Event struct {
	Reason        string      `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason"`
	Message       string      `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`
	LastTimestamp metav1.Time `json:"lastTimestamp,omitempty" protobuf:"bytes,7,opt,name=lastTimestamp"`
	Type          string      `json:"type,omitempty" protobuf:"bytes,9,opt,name=type"`
}
