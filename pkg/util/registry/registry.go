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

package registry

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/docker/distribution/manifest/schema1"
)

var (
	insecureHttpTransport, secureHttpTransport *http.Transport
)

type Registry struct {
	Endpoint *url.URL
	Client   *http.Client
}

func init() {
	secureHttpTransport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}
	insecureHttpTransport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
}

func GetHttpTransport(insecure bool) *http.Transport {
	if insecure {
		return insecureHttpTransport
	}
	return secureHttpTransport
}

// FormatEndpoint formats endpoint
func FormatEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	endpoint = strings.TrimRight(endpoint, "/")
	if !strings.HasPrefix(endpoint, "http://") &&
		!strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}

	return endpoint
}

// ParseEndpoint parses endpoint to a URL
func ParseEndpoint(endpoint string) (*url.URL, error) {
	endpoint = FormatEndpoint(endpoint)

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func NewRegistry(endpoint string, client *http.Client) (*Registry, error) {
	url, err := ParseEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	return &Registry{
		Endpoint: url,
		Client:   client,
	}, nil
}

func (this *Registry) GetCatalog() ([]string, error) {
	repos := []string{}
	suffix := "/v2/_catalog"
	url := fmt.Sprintf("%s%s", this.Endpoint.String(), suffix)
	res, err := this.Client.Get(url)
	if err != nil {
		return repos, err
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return repos, err
	}

	if res.StatusCode == http.StatusOK {
		cataLogs := &struct {
			Repositories []string `json:repositories`
		}{[]string{}}
		if err = json.Unmarshal(data, cataLogs); err != nil {
			return repos, err
		}
		repos = cataLogs.Repositories
	} else {
		return repos, errors.New(string(data))
	}

	return repos, nil
}

func (this *Registry) GetTags(catalog string) ([]string, error) {
	tags := []string{}
	suffix := fmt.Sprintf("/v2/%s/tags/list", catalog)
	url := fmt.Sprintf("%s%s", this.Endpoint.String(), suffix)
	res, err := this.Client.Get(url)
	if err != nil {
		return []string{}, err
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return tags, err
	}

	if res.StatusCode == http.StatusOK {
		tag := &struct {
			Tags []string `json:tags`
		}{[]string{}}
		if err = json.Unmarshal(data, &tag); err != nil {
			return tags, err
		}
		tags = tag.Tags
	} else {
		return tags, errors.New(string(data))
	}
	return tags, nil
}

func (this *Registry) GetManifest(catalog, reference string) (*schema1.Manifest, error) {
	manifest := new(schema1.Manifest)
	suffix := fmt.Sprintf("/v2/%s/manifests/%s", catalog, reference)
	url := fmt.Sprintf("%s%s", this.Endpoint.String(), suffix)
	res, err := this.Client.Get(url)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusOK {
		if err = json.Unmarshal(data, manifest); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New(string(data))
	}
	return manifest, nil
}

func (this *Registry) DelManifest(catalog, reference string) error {
	suffix := fmt.Sprintf("/v2/%s/manifests/%s", catalog, reference)
	url := fmt.Sprintf("%s%s", this.Endpoint.String(), suffix)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	res, err := this.Client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == 202 {
		return nil
	} else if res.StatusCode == 404 {
		return errors.New(fmt.Sprintf("the image named %v not found", reference))
	}
	return err
}
