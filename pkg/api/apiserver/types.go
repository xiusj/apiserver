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

package apiserver

import (
	"time"
)

type App struct {
	ID            uint       `json:"id"`
	CreatedAt     time.Time  `json:"createAt"`
	Name          string     `json:"nmae,omitempty"`
	UserName      string     `json:"nameSpace,omitempty"`
	Description   string     `json:"description,omitempty"`
	AppStatus     int        `json:"appStatus,omitempty"`
	ServiceCount  int        `json:"serviceCount,omitempty"`
	InstanceCount int        `json:"intanceCount,omitempty"`
	External      string     `json:"external,omitempty"`
	Items         []*Service `json:"services,omitempty"`
}

type Service struct {
	ID            uint           `json:"id"`
	CreatedAt     time.Time      `json:"createAt"`
	Name          string         `json:"name,omitempty"`
	Image         string         `json:"image,omitempty"`
	InstanceCount int            `json:"instanceCount,omitempty" `
	Status        int            `json:"status,omitempty"`
	Type          int            `json:"type,omitempty"` //0 stateless 1 statefulset
	External      string         `json:"external,omitempty"`
	LoadbalanceIp string         `json:"loadbalanceIp,omitempty"`
	Config        *ServiceConfig `json:"config,omitempty"`
	Storage       *Storage       `json:"storage,omitempty"`
	AppName       string         `json:"appName,omitempty"`
	Items         []*Container   `json:"containers,omitempty"`
	AppId         uint           `json:"appId,omitempty"`
}

type Container struct {
	ID        uint             `json:"id"`
	CreatedAt time.Time        `json:"createAt"`
	Name      string           `json:"name,omitempty"`
	Image     string           `json:"image,omitempty"`
	Status    int              `json:"status,omitempty"`
	Internal  string           `json:"internal,omitempty"`
	AppName   string           `json:"appName,omitempty"`
	Config    *ContainerConfig `json:"config,omitempty"`
	ServiceId uint
}

type Storage struct {
	Type string `json:"Type"`
	Size string `json:"Size"`
}

type ServiceConfig struct {
	ID          uint         `json:"id"`
	CreatedAt   time.Time    `json:"createAt"`
	BaseConfig  *BaseConfig  `json:"base,omitempty"`
	ConfigGroup *ConfigGroup `json:"configGroup,omitempty"`
	SuperConfig *SuperConfig `json:"super,omitempty"`
	ServiceId   uint
}

type ContainerConfig struct {
	ID          uint         `json:"id"`
	CreatedAt   time.Time    `json:"createAt"`
	BaseConfig  *BaseConfig  `json:"base,omitempty"`
	ConfigGroup *ConfigGroup `json:"configGroup,omitempty"`
	SuperConfig *SuperConfig `json:"super,omitempty"`
	ContainerId uint
}

type BaseConfig struct {
	ID              uint      `json:"id"`
	CreatedAt       time.Time `json:"createAt"`
	Cpu             string    `json:"cpu,omitempty"`
	Memory          string    `json:"memory,omitempty"`
	Type            int       `json:"type,omitempty"` //0 stateless 1 stateful
	Volumes         []*Volume `json:"volumes,omitempty"`
	ServiceConfigId uint
}

type Volume struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"createAt"`
	TargetPath   string    `json:"targetPath,omitempty"`
	Storage      string    `json:"storage,omitempty"`
	BaseConfigId uint
}

type ConfigGroup struct {
	ID              uint         `json:"id"`
	CreatedAt       time.Time    `json:"createAt"`
	Namespace       string       `json:"namespace"`
	Name            string       `json:"name,omitempty"`
	ServiceName     string       `json:"serviceName,omitempty"`
	ConfigMaps      []*ConfigMap `json:"items,omitempty"`
	ServiceConfigId uint
}

type ConfigMap struct {
	ID            uint      `json:"id"`
	CreatedAt     time.Time `json:"createAt"`
	Name          string    `json:"name,omitempty" `
	Content       string    `json:"content,omitempty"`
	ContainerPath string    `json:"containerPath,omitempty"`
	ConfigGroupId uint
}

type SuperConfig struct {
	ID              uint      `json:"id"`
	CreatedAt       time.Time `json:"createAt"`
	Envs            []*Env    `json:"envs,omitempty"`
	Ports           []*Port   `json:"ports,omitempty"`
	ServiceConfigId uint
}

type Env struct {
	ID            uint      `json:"id"`
	CreatedAt     time.Time `json:"createAt"`
	Key           string    `json:"key,omitempty"`
	Val           string    `json:"val,omitempty"`
	SuperConfigId uint
}

type Port struct {
	ID            uint      `json:"id"`
	CreatedAt     time.Time `json:"createAt"`
	ContainerPort int       `json:"containerPort,omitempty"`
	ServicePort   int       `json:"servicePort,omitempty"`
	Protocol      string    `json:"protocol,omitempty"`
	SuperConfigId uint
}

type ScaleOption struct {
	ServiceInstanceCnt int `json:"serviceInstanceCnt"`
}

type ExpansionOption struct {
	Cpu    string `json:"cpu"`
	Memory string `json:"memory"`
}

type RollOption struct {
	Image  string       `json:"image"`
	Conifg *ConfigGroup `json:"config"`
}

type Deploy struct {
	ID              uint          `json:"requirementId,omitempty"`
	requirementName string        `json:"requirementName,omitempty"`
	Type            string        `json:"type,omitempty"` //previewDeploy,productDeploy,rollBack
	Items           []*DeployItem `json:"features,omitempty"`
}

type DeployItem struct {
	ID            uint   `json:"id,omitempty"`
	ProjectId     uint   `projectId:"id,omitempty"`
	ProjectName   string `json:"projectName,omitempty"`
	DockerRepoUrl string `json:"dockerRepoUrl,omitempty"`
	Tag           string `json:"tag,omitempty"`
	Status        int    `json:"type,omtempty"` //0 not deploy 1 deployed
	DeployId      uint   `json:"deployId,omitempty"`
}

type ProjectConfig struct {
	ID        uint      `json:"id,omitempty"`
	ProjectId uint      `json:"projectId,omitempty"`
	Key       string    `json:"key,omitempty"`
	Val       string    `json:"val,omitempty"`
	Type      string    `json:"type,omitempty"`
	CreateAt  time.Time `json:"createAt,omitempty"`
	UpdateAt  time.Time `json:"modifyAt,omitempty"`
	Operator  string    `json:"operator,omitempty"`
}

type ProjectConfigOption struct {
	ProjectId uint   `json:"projectId,omitempty"`
	Key       string `json:"key,omitempty"`
	Val       string `json:"val,omitempty"`
	Type      string `json:"type,omitempty"`
	CreateAt  string `json:"createAt,omitempty"`
	UpdateAt  string `json:"modifyAt,omitempty"`
	Operator  string `json:"operator,omitempty"`
}

type Result struct {
	ID             uint          `json:"requirementId,omitempty"`
	CallbackResult string        `json:"callbackResult,omityempty"` //SUCCESS,FAILURE,UNKNOW
	CallbackType   string        `json:"callbackType,omitempty"`    //previewDeploy,productDeploy,rollBack
	Operator       string        `json:"operator,omitempty"`
	Items          []*ResultItem `json:"featureResultDtoList,omitempty"`
}

type ResultItem struct {
	ID             uint   `json:"id,omitempty"`
	CurrentVersion string `json:"currentVersion,omitempty"`
	Status         string `json:"status,omityempty"` //SUCCESS,FAILURE,UNKNOW
}

type Logs struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createAt"`
	UserName  string    `json:",omitempty"`
	AppName   string    `json:",omitempty"`
	EventType string    `json:",omitempty"`
}

//Process container's process
type Process struct {
	User        string  `json:"user"`
	PID         int64   `json:"pid"`
	ParentPID   int64   `json:"parent_pid"`
	StartTime   string  `json:"start_time"`
	PercentCPU  float64 `json:"percent_cpu"`
	PercentMEM  float64 `json:"percent_mem"`
	rss         int64   `json:"rss"`
	VirtualSize int64   `json:"virtual_size"`
	Status      string  `json:"status"`
	RunningTime string  `json:"running_time"`
	CgroupPath  string  `json:"cgroup_path"`
	Cmd         string  `json:"cmd"`
}

type Item struct {
	Name string `json:"name"`
}

type CadvisorResult struct {
	Subcontainers []*Item `json:"subcontainers"`
}
