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

package cache

import (
	"fmt"
	"sync"
	"time"

	"apiserver/pkg/api/apiserver"
	"apiserver/pkg/client"
	"apiserver/pkg/resource"
	"apiserver/pkg/util/log"

	betav1 "k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	MAX_SIZE          = 10240
	LIST_WATCH_PERIOD = 30
)

var (
	Store     *Cache
	firstSync = true
)

type Cache struct {
	*NamespaceCache
	*ServiceCache
	*StatefulSetCache
	*DeploymentCache
	*ConfigMapCache
	*PodCache
}

type NamespaceCache struct {
	sync.RWMutex
	List map[string]map[string]v1.Namespace
}

type ServiceCache struct {
	sync.RWMutex
	List map[string]map[string]v1.Service
}

type StatefulSetCache struct {
	sync.RWMutex
	List map[string]map[string]betav1.StatefulSet
}

type DeploymentCache struct {
	sync.RWMutex
	List map[string]map[string]extensions.Deployment
}

type ConfigMapCache struct {
	sync.RWMutex
	List map[string]map[string]v1.ConfigMap
}

type PodCache struct {
	sync.RWMutex
	List map[string]map[string]v1.Pod
}

func init() {
	Store = &Cache{
		&NamespaceCache{
			List: make(map[string]map[string]v1.Namespace, MAX_SIZE),
		},
		&ServiceCache{
			List: make(map[string]map[string]v1.Service, MAX_SIZE),
		},
		&StatefulSetCache{
			List: make(map[string]map[string]betav1.StatefulSet, MAX_SIZE),
		},
		&DeploymentCache{
			List: make(map[string]map[string]extensions.Deployment, MAX_SIZE),
		},
		&ConfigMapCache{
			List: make(map[string]map[string]v1.ConfigMap, MAX_SIZE),
		},
		&PodCache{
			List: make(map[string]map[string]v1.Pod, MAX_SIZE),
		},
	}
}

//List sync ervery 30 Second to list k8s's resource to memory
func List() {
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ticker.C:
			go ListResource()
		}
	}
}

//Watch sync ervery 30 Second to update the cntainer's status or insert container info to local database
func Watch() {
	ticker := time.NewTicker(time.Second * 15)
	for {
		select {
		case <-ticker.C:
			go UpdateAppStatus()
		}
	}
}

func UpdateAppStatus() {
	for k, _ := range Store.NamespaceCache.List {
		podlist := Store.PodCache.List[k]
		apps := apiserver.QueryAppsByNamespace(k)
		for _, app := range apps {
			svcs := apiserver.QueryServicesByAppId(app.ID)
			for _, svc := range svcs {
				for _, pod := range podlist {
					if svc.Name == pod.ObjectMeta.Labels["name"] {
						container := &apiserver.Container{
							Name:      pod.ObjectMeta.Name,
							Image:     pod.Spec.Containers[0].Image,
							Internal:  fmt.Sprintf("%v:%v", pod.Status.PodIP, pod.Spec.Containers[0].Ports[0].HostPort),
							ServiceId: svc.ID,
						}
						if pod.Status.Phase == "Running" {
							container.Status = resource.AppRunning
						}
						if pod.Status.Phase == "Pending" {
							container.Status = resource.AppBuilding
						}
						if pod.Status.Phase == "Succeeded" {
							container.Status = resource.AppSuccessed
						}
						if pod.Status.Phase == "Failed" {
							container.Status = resource.AppFailed
						}
						if pod.Status.Phase == "Unknown" {
							container.Status = resource.AppUnknow
						}

						if apiserver.ExistContainer(&apiserver.Container{Name: pod.ObjectMeta.Name}) {
							apiserver.InsertContainer(container)
						} else {
							c, _ := apiserver.QueryContainerByName(pod.ObjectMeta.Name)
							container.ID = c.ID
							apiserver.UpdateContainer(container)
						}
						svc.Status = container.Status
					}
				}
				apiserver.UpdateServiceOnly(svc)
				app.AppStatus = svc.Status
			}
			apiserver.UpdateAppOnly(app)
		}
	}
}

//watch and list k8s's resource (namespace,service,replicationController) to resource memory
func ListResource() {
	nsList, err := client.Client.K8sClient.
		CoreV1().
		Namespaces().
		List(metav1.ListOptions{})
	if err != nil {
		log.Errorf("list and watch k8s's namespace err: %v", err)
		return
	} else {
		if len(nsList.Items) > 0 {
			loop(nsList, "")
		}
	}
	for k, v := range Store.NamespaceCache.List {
		svcList, err := client.Client.K8sClient.
			CoreV1().
			Services(v[k].ObjectMeta.Name).
			List(metav1.ListOptions{})
		if err != nil {
			log.Errorf("list and watch k8s's service of namespace [%v] err: %v", v[k].Name, err)
		} else {
			loop(svcList, v[k].ObjectMeta.Name)
		}

		stfList, err := client.Client.K8sClient.
			StatefulSets(v[k].ObjectMeta.Name).
			List(metav1.ListOptions{})
		if err != nil {
			log.Errorf("list and watch k8s's stfList of namespace [%v] err: %v", v[k].Name, err)
		} else {
			loop(stfList, v[k].ObjectMeta.Name)
		}

		dpList, err := client.Client.K8sClient.
			ExtensionsV1beta1().
			Deployments(v[k].ObjectMeta.Name).
			List(metav1.ListOptions{})
		if err != nil {
			log.Errorf("list and watch k8s's deployment of namespace [%v] err: %v", v[k].Name, err)
		} else {
			loop(dpList, v[k].ObjectMeta.Name)
		}

		cfgMapList, err := client.Client.K8sClient.
			CoreV1().
			ConfigMaps(v[k].ObjectMeta.Name).
			List(metav1.ListOptions{})
		if err != nil {
			log.Errorf("list and watch k8s's configMap of namespace [%v] err: %v", v[k].Name, err)
		} else {
			loop(cfgMapList, v[k].ObjectMeta.Name)
		}

		podList, err := client.Client.K8sClient.
			CoreV1().
			Pods(v[k].ObjectMeta.Name).
			List(metav1.ListOptions{})
		if err != nil {
			log.Errorf("list and watch k8s's configMap of namespace [%v] err: %v", v[k].Name, err)
		} else {
			loop(podList, v[k].ObjectMeta.Name)
		}
	}
}

//loop add the k8s's resource (namespace,service,replicationController) to resource map
func loop(param interface{}, nsname string) {
	switch param.(type) {
	case *v1.NamespaceList:
		for _, ns := range param.(*v1.NamespaceList).Items {
			if ns.Name == "kube-system" || ns.Name == "kube-public" || ns.Name == "default" {
				continue
			}
			Store.NamespaceCache.Lock()
			nsmap := make(map[string]v1.Namespace)
			nsmap[ns.ObjectMeta.Name] = ns
			Store.NamespaceCache.List[ns.ObjectMeta.Name] = nsmap
			Store.NamespaceCache.Unlock()
		}
	case *v1.ServiceList:
		items := param.(*v1.ServiceList).Items
		if len(items) == 0 {
			Store.ServiceCache.List[nsname] = make(map[string]v1.Service)
		} else {
			Store.ServiceCache.Lock()
			svcmap := make(map[string]v1.Service)
			for _, svc := range items {
				svcmap[svc.ObjectMeta.Name] = svc
			}
			Store.ServiceCache.List[nsname] = svcmap
			Store.ServiceCache.Unlock()
		}
	case *betav1.StatefulSetList:
		items := param.(*betav1.StatefulSetList).Items
		if len(items) == 0 {
			Store.StatefulSetCache.List[nsname] = make(map[string]betav1.StatefulSet)
		} else {
			Store.StatefulSetCache.Lock()
			stfpmap := make(map[string]betav1.StatefulSet)
			for _, statefulSet := range items {
				stfpmap[statefulSet.ObjectMeta.Name] = statefulSet
			}
			Store.StatefulSetCache.List[nsname] = stfpmap
			Store.StatefulSetCache.Unlock()
		}
	case *extensions.DeploymentList:
		items := param.(*extensions.DeploymentList).Items
		if len(items) == 0 {
			Store.DeploymentCache.List[nsname] = make(map[string]extensions.Deployment)
		} else {
			Store.DeploymentCache.Lock()
			dpmap := make(map[string]extensions.Deployment)
			for _, deploy := range items {
				dpmap[deploy.ObjectMeta.Name] = deploy
			}
			Store.DeploymentCache.List[nsname] = dpmap
			Store.DeploymentCache.Unlock()
		}
	case *v1.ConfigMapList:
		items := param.(*v1.ConfigMapList).Items
		if len(items) == 0 {
			Store.ConfigMapCache.List[nsname] = make(map[string]v1.ConfigMap)
		} else {
			Store.ConfigMapCache.Lock()
			cfgmap := make(map[string]v1.ConfigMap, MAX_SIZE)
			for _, configMap := range items {
				cfgmap[configMap.ObjectMeta.Name] = configMap
			}
			Store.ConfigMapCache.List[nsname] = cfgmap
			Store.ConfigMapCache.Unlock()
		}
	case *v1.PodList:
		items := param.(*v1.PodList).Items
		if len(items) == 0 {
			Store.PodCache.List[nsname] = make(map[string]v1.Pod)
		} else {
			Store.PodCache.Lock()
			podMap := make(map[string]v1.Pod, MAX_SIZE)
			for _, pod := range items {
				if pod.ObjectMeta.DeletionTimestamp == nil {
					podMap[pod.ObjectMeta.Name] = pod
				}
			}
			Store.PodCache.List[nsname] = podMap
			Store.PodCache.Unlock()
		}
	}
}

//ExsitResource decide namespace,service,replicationController,deployment,pod,configMap of k8s resource exsit or not by name;false is not exsit,true exsit
func ExsitResource(namespace, resourceName, resourceKind string) bool {
	switch resourceKind {
	case resource.ResourceKindNamespace:
		_, exist := Store.NamespaceCache.List[resourceName]
		return exist
	case resource.ResourceKindService:
		_, exist := Store.ServiceCache.List[namespace][resourceName]
		return exist
	case resource.ResourceKindStatefulSet:
		_, exist := Store.StatefulSetCache.List[namespace][resourceName]
		return exist
	case resource.ResourceKindDeployment:
		_, exist := Store.DeploymentCache.List[namespace][resourceName]
		return exist
	case resource.ResourceKindConfigMap:
		_, exist := Store.ConfigMapCache.List[namespace][resourceName]
		return exist
	}
	return false
}
