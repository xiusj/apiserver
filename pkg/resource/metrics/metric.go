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

package metrics

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"apiserver/pkg/client"
	"apiserver/pkg/configz"

	"apiserver/pkg/util/log"
)

func GetPodMetrics(namespace, podName, metric_name string) (map[string]interface{}, error) {
	path := fmt.Sprintf("%s/api/v1/model/namespaces/%s/pods/%s/metrics/%s", configz.GetString("apiserver", "heapsterEndpoint", "127.0.0.1:30003"), namespace, podName, metric_name)
	log.Debug(path)
	heapsterHost := configz.GetString("apiserver", "heapsterEndpoint", "http://127.0.0.1:30003")
	log.Infof("Creating remote Heapster client for %s", heapsterHost)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Client.Heapsterclient.Do(req)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(res.Body)
	v := map[string]interface{}{}
	json.Unmarshal(data, &v)
	return v, nil
}
