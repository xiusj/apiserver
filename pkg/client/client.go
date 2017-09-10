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

package client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"apiserver/pkg/api/apiserver"
	"apiserver/pkg/configz"
	"apiserver/pkg/resource"
	"apiserver/pkg/resource/event"
	"apiserver/pkg/util/log"
	"apiserver/pkg/util/parseUtil"

	docker "github.com/docker/docker/client"

	betav1 "k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Handler interface {
	CreateResource(param interface{}) error
	ExsitResource(param interface{}) bool
	DeleteResource(param interface{}) error
	UpdateResouce(param interface{}) error
}

var Client *Clientset

type Clientset struct {
	K8sClient      *k8s.Clientset
	DockerClient   *docker.Client
	Heapsterclient *http.Client
	RegistryClient *http.Client
	K8sConfig      *restclient.Config
}

//init initialize the clientset
func init() {
	//create k8s clientset
	config, err := clientcmd.BuildConfigFromFlags("", configz.GetString("apiserver", "k8s-config", "./config"))
	if err != nil {
		log.Fatalf("init k8s config err: %v", err)
	}
	k8sClient, err := k8s.NewForConfig(config)
	if err != nil {
		log.Fatalf("init k8s client err: %v", err)
	}

	//create docker client
	host := configz.GetString("build", "endpoint", "127.0.0.1:2375")
	version := configz.GetString("build", "version", "1.24")
	cl := &http.Client{
		Transport: new(http.Transport),
	}
	dockerClient, err := docker.NewClient(host, version, cl, nil)
	if err != nil {
		log.Fatalf("init docker client err: %v", err)
	}

	//create Clientset
	Client = &Clientset{K8sClient: k8sClient, DockerClient: dockerClient, Heapsterclient: http.DefaultClient, RegistryClient: http.DefaultClient, K8sConfig: config}
}

//NewClientset return Clientset
func NewClientset(k8sconfig, dockerEndpoint, dockerAPIVersion string) (*Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", k8sconfig)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("init k8s config err: %v", err))
	}
	k8sClient, err := k8s.NewForConfig(config)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("init k8s client err: %v", err))
	}

	//create docker client
	cl := &http.Client{
		Transport: new(http.Transport),
	}
	dockerClient, err := docker.NewClient(dockerEndpoint, dockerAPIVersion, cl, nil)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("init docker client err: %v", err))
	}

	//create Clientset
	return &Clientset{K8sClient: k8sClient, DockerClient: dockerClient, Heapsterclient: http.DefaultClient, RegistryClient: http.DefaultClient}, nil
}

//CreateResource create namespace,service,replicationController,deployment,pod,configMa of k8s resource
func (client *Clientset) CreateResource(param interface{}) error {
	switch param.(type) {
	case *v1.Namespace:
		ns := param.(*v1.Namespace)
		_, err := client.K8sClient.
			CoreV1().
			Namespaces().
			Create(ns)
		if err != nil {
			log.Errorf("create namespace [%v] err:%v", ns.Name, err)
			return err
		}
		log.Noticef("namespace [%v] is created]", ns.Name)
		return nil
	case *v1.Service:
		svc := param.(*v1.Service)
		_, err := client.K8sClient.
			CoreV1().
			Services(svc.Namespace).
			Create(svc)
		if err != nil {
			log.Errorf("create service [%v] err:%v", svc.Name, err)
			return err
		}
		log.Noticef("service [%v] is created]", svc.Name)
		return nil
	case *v1.ConfigMap:
		cfgMap := param.(*v1.ConfigMap)
		_, err := client.K8sClient.
			CoreV1().
			ConfigMaps(cfgMap.Namespace).
			Create(cfgMap)
		if err != nil {
			log.Errorf("create configMap [%v] err:%v", cfgMap.Name, err)
			return err
		}
		log.Noticef("configMap [%v] is created]", cfgMap.Name)
		return nil
	case *v1.ReplicationController:
		rc := param.(*v1.ReplicationController)
		_, err := client.K8sClient.
			CoreV1().
			ReplicationControllers(rc.Namespace).
			Create(rc)
		if err != nil {
			log.Errorf("create replicationControllers [%v] err:%v", rc.Name, err)
			return err
		}
		log.Noticef("replication [%v] is created]", rc.Name)
		return nil
	case *extensions.Deployment:
		deploy := param.(*extensions.Deployment)
		_, err := client.K8sClient.ExtensionsV1beta1().Deployments(deploy.Namespace).Create(deploy)
		if err != nil {
			log.Errorf("create deployment [%v] err:%v", deploy.Name, err)
			return err
		}
		log.Noticef("deployment [%v] is created]", deploy.Name)
		return nil

	case *betav1.StatefulSet:
		statefulSet := param.(*betav1.StatefulSet)
		_, err := client.K8sClient.StatefulSets(statefulSet.Namespace).Create(statefulSet)
		if err != nil {
			log.Errorf("create statefulSet [%v] err:%v", statefulSet.Name, err)
			return err
		}
		log.Noticef("statefulSet [%v] is created]", statefulSet.Name)
		return nil
	}

	return nil
}

//ExsitResource decide namespace,service,replicationController,deployment,pod,configMap of k8s resource exsit or not by name;false is not exsit,true exsit
func (client *Clientset) ExsitResource(param interface{}) bool {
	switch param.(type) {
	case v1.Namespace:
		_, err := client.K8sClient.
			CoreV1().
			Namespaces().
			Get(param.(v1.Namespace).Name, metav1.GetOptions{})
		if err != nil {
			return false
		}
		return true
	case v1.Service:
		svc := param.(v1.Service)
		_, err := client.K8sClient.
			CoreV1().
			Services(svc.Namespace).
			Get(svc.Name, metav1.GetOptions{})
		if err != nil {
			return false
		}
		return true

	case v1.ConfigMap:
		cfgMap := param.(v1.Service)
		_, err := client.K8sClient.
			CoreV1().
			ConfigMaps(cfgMap.Namespace).
			Get(cfgMap.Name, metav1.GetOptions{})
		if err != nil {
			return false
		}
		return true
	case v1.ReplicationController:
		rc := param.(v1.ReplicationController)
		_, err := client.K8sClient.
			CoreV1().
			ReplicationControllers(rc.Namespace).
			Get(rc.Name, metav1.GetOptions{})
		if err != nil {
			return false
		}
		return true
	case extensions.Deployment:
		deploy := param.(*extensions.Deployment)
		_, err := client.K8sClient.
			ExtensionsV1beta1().
			Deployments(deploy.Namespace).
			Get(deploy.Name, metav1.GetOptions{})
		if err != nil {
			return false
		}
		return true
	case *betav1.StatefulSet:
		statefulSet := param.(*betav1.StatefulSet)
		_, err := client.K8sClient.StatefulSets(statefulSet.Namespace).Get(statefulSet.Name, metav1.GetOptions{})
		if err != nil {
			return false
		}
		return true
	}
	return false
}

//DeleteResource delete namespace,service,replicationController,deployment,pod,configMap of k8s resource
func (client *Clientset) DeleteResource(param interface{}) error {
	backend := new(metav1.DeletionPropagation)
	*backend = metav1.DeletePropagationForeground
	switch param.(type) {
	case v1.Namespace:
		ns := param.(v1.Namespace)
		err := client.K8sClient.
			CoreV1().
			Namespaces().
			Delete(ns.Name, &metav1.DeleteOptions{TypeMeta: resource.NewTypeMeta(resource.ResourceKindNamespace, "v1"), GracePeriodSeconds: parseUtil.IntToInt64Pointer(30)})
		if err != nil {
			log.Errorf("delete namespace [%v] err:%v", ns.Name, err)
			return err
		}
		log.Noticef("namespace [%v] was deleted", ns.Name)
		return nil
	case v1.Service:
		svc := param.(v1.Service)
		err := client.K8sClient.
			CoreV1().
			Services(svc.Namespace).
			Delete(svc.Name, &metav1.DeleteOptions{})
		if err != nil {
			log.Errorf("delete service [%v] err:%v", svc.Name, err)
			return err
		}
		log.Noticef("service [%v] was deleted]", svc.Name)
		return nil

	case v1.ConfigMap:
		cfgMap := param.(v1.ConfigMap)
		err := client.K8sClient.
			CoreV1().
			ConfigMaps(cfgMap.Namespace).
			Delete(cfgMap.Name, &metav1.DeleteOptions{})
		if err != nil {
			log.Errorf("delete configMap [%v] err:%v", cfgMap.Name, err)
			return err
		}
		log.Noticef("configMap [%v] was deleted]", cfgMap.Name)
		return nil
	case v1.ReplicationController:
		rc := param.(v1.ReplicationController)
		err := client.K8sClient.
			CoreV1().
			ReplicationControllers(rc.Namespace).
			Delete(rc.Name, &metav1.DeleteOptions{TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "ReplicationController"}, OrphanDependents: parseUtil.BoolToPointer(false)})
		if err != nil {
			log.Errorf("delete replicationControllers [%v] err:%v", rc.Name, err)
			return err
		}
		log.Noticef("replication [%v] is delete]", rc.Name)
		return nil
	case extensions.Deployment:
		deploy := param.(extensions.Deployment)
		err := client.K8sClient.
			ExtensionsV1beta1().
			Deployments(deploy.Namespace).
			Delete(deploy.Name, &metav1.DeleteOptions{PropagationPolicy: backend})
		if err != nil {
			log.Errorf("delete deployment [%v] err:%v", deploy.Name, err)
			return err
		}
		log.Noticef("deployment [%v] was deleted]", deploy.Name)
		return nil
	case v1.Pod:
		pod := param.(v1.Pod)
		err := client.K8sClient.
			CoreV1().
			Pods(pod.Namespace).
			Delete(pod.Name, &metav1.DeleteOptions{PropagationPolicy: backend})
		if err != nil {
			log.Errorf("delete pod [%v] err:%v", pod.Name, err)
			return err
		}
		log.Noticef("pod [%v] was deleted]", pod.Name)
		return nil
	case betav1.StatefulSet:
		statefulSet := param.(betav1.StatefulSet)
		err := client.K8sClient.
			StatefulSets(statefulSet.Namespace).
			Delete(statefulSet.Name, &metav1.DeleteOptions{PropagationPolicy: backend})
		if err != nil {
			log.Errorf("delete statefulSet [%v] err:%v", statefulSet.Name, err)
			return err
		}
		log.Noticef("statefulSet [%v] was deleted]", statefulSet.Name)
		return nil
	}
	return nil
}

//UpdateResouce update namespace,service,replicationController,deployment,pod,configMap of k8s resource
func (client *Clientset) UpdateResouce(param interface{}) error {
	switch param.(type) {
	case *v1.Namespace:
		ns := param.(*v1.Namespace)
		_, err := client.K8sClient.
			CoreV1().
			Namespaces().
			Update(ns)
		if err != nil {
			log.Errorf("update namespace [%v] err:%v", ns.Name, err)
			return err
		}
		log.Noticef("namespace [%v] was updated", ns.Name)
		return nil
	case *v1.Service:
		svc := param.(*v1.Service)
		_, err := client.K8sClient.
			CoreV1().
			Services(svc.Namespace).
			Update(svc)
		if err != nil {
			log.Errorf("update service [%v] err:%v", svc.Name, err)
			return err
		}
		log.Noticef("service [%v] was updated]", svc.Name)
		return nil
	case *v1.ReplicationController:
		rc := param.(*v1.ReplicationController)
		_, err := client.K8sClient.
			CoreV1().
			ReplicationControllers(rc.Namespace).
			Update(rc)
		if err != nil {
			log.Errorf("update replicationControllers [%v] err:%v", rc.Name, err)
			return err
		}
		log.Noticef("replication [%v] is updated]", rc.Name)
		return nil

	case *extensions.Deployment:
		deploy := param.(*extensions.Deployment)
		_, err := client.K8sClient.
			ExtensionsV1beta1().
			Deployments(deploy.Namespace).Update(deploy)
		if err != nil {
			log.Errorf("update replicationControllers [%v] err:%v", deploy.Name, err)
			return err
		}
		log.Noticef("replication [%v] is updated]", deploy.Name)
		return nil
	case *v1.ConfigMap:
		cfgMap := param.(*v1.ConfigMap)
		_, err := client.K8sClient.CoreV1().ConfigMaps(cfgMap.Namespace).Update(cfgMap)
		if err != nil {
			log.Errorf("update configMap [%v] err:%v", cfgMap.Name, err)
			return err
		}
		log.Noticef("configMap [%v] is updated]", cfgMap.Name)
		return nil
	case *betav1.StatefulSet:
		statefulSet := param.(*betav1.StatefulSet)
		_, err := client.K8sClient.
			StatefulSets(statefulSet.Namespace).
			Update(statefulSet)
		if err != nil {
			log.Errorf("update statefulSet [%v] err:%v", statefulSet.Name, err)
			return err
		}
		log.Noticef("statefulSet [%v] is updated]", statefulSet.Name)
		return nil
	}
	return nil
}

//GetPods return the pods of the give namespace and deployment's name
func (client *Clientset) GetPods(namespace, deployName string) ([]v1.Pod, error) {
	list, err := client.K8sClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "name=" + deployName})
	if err != nil {
		return []v1.Pod{}, err
	}
	return list.Items, nil
}

//CreateService create k8s service by service
func (client *Clientset) CreateService(svc *v1.Service) (*v1.Service, error) {
	service, err := client.K8sClient.
		CoreV1().
		Services(svc.Namespace).
		Create(svc)
	if err != nil {
		log.Errorf("create service [%v] err:%v", svc.Name, err)
		return nil, err
	}
	log.Noticef("service [%v] was created]", svc.Name)
	return service, nil
}

//GetDeploymentPods return deployment's pods by namespace and deployment's name
func (client *Clientset) GetDeploymentPods(appName, namespace string) ([]v1.Pod, error) {
	list, err := client.K8sClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "name=" + appName})
	if err != nil {
		log.Errorf("get deployment [%v]'s pods err:%v", appName, err)
		return nil, err
	}
	return list.Items, nil
}

//GetEventsForContainer return pod's event by namespace and pod's name
func (client *Clientset) GetEventsForContainer(namespace, containerName string) (list []event.Event, err error) {
	listEvent, err := client.K8sClient.CoreV1().Events(namespace).List(resource.ListEverything)
	if err != nil {
		log.Errorf("get pod [%v]'s event err:%v", containerName, err)
		return
	}
	for _, ev := range listEvent.Items {
		if strings.Contains(ev.Name, containerName) {
			list = append(
				list,
				event.Event{
					Reason:        ev.Reason,
					Type:          ev.Type,
					LastTimestamp: ev.LastTimestamp,
					Message:       ev.Message,
				},
			)
		}
	}
	return
}

//GetEventsForService return service's event by namespace and service's name
func (client *Clientset) GetEventsForService(namespace, serviceName string) (list []event.Event, err error) {
	listEvent, err := client.K8sClient.CoreV1().Events(namespace).List(resource.ListEverything)
	if err != nil {
		log.Errorf("get service [%v]'s event err:%v", serviceName, err)
		return
	}

	svcs, _ := apiserver.QueryServices(serviceName, 100000, 0, 0)
	if len(svcs[0].Items) > 0 {
		for _, ev := range listEvent.Items {
			if strings.Contains(ev.Name, svcs[0].Items[0].Name) {
				list = append(
					list,
					event.Event{
						Reason:        ev.Reason,
						Type:          ev.Type,
						LastTimestamp: ev.LastTimestamp,
						Message:       ev.Message,
					},
				)
			}

			if strings.Contains(ev.Name, serviceName) {
				list = append(
					list,
					event.Event{
						Reason:        ev.Reason,
						Type:          ev.Type,
						LastTimestamp: ev.LastTimestamp,
						Message:       ev.Message,
					},
				)
			}
		}

	}
	return
}

//GetLogForContainer return pod's log by namespace and pod's name
func (client *Clientset) GetLogForContainer(namespace, podName string, logOptions *v1.PodLogOptions) (string, error) {
	req := client.K8sClient.CoreV1().RESTClient().Get().
		Namespace(namespace).
		Name(podName).
		Resource("pods").
		SubResource("log").
		VersionedParams(logOptions, scheme.ParameterCodec)

	readCloser, err := req.Stream()
	if err != nil {
		return err.Error(), nil
	}

	defer func() {
		if err = readCloser.Close(); err != nil {
			log.Errorf("close readstream err:%v", err)
		}
	}()
	result, err := ioutil.ReadAll(readCloser)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

func (client *Clientset) ExecContainer() {

}
