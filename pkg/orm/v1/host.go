package v1

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
)

type Host struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	Spec            HostSpec
}

type HostSpec struct {
	SSH     HostSSH
	Info    HostInfo
	Plugins []HostPlugin
	Sdks    []SdkPlugin
}

type HostSSH struct {
	Host     string
	User     string
	Password string
	Port     uint16
}

type HostInfo struct {
	OS     OS
	CPU    CPU
	Memory Memory
	Disk   Disk
	GPUs   []GPUInfo
}

type OS struct {
	Release string
	Kernel  string
}

type CPU struct {
	Cores int
	Model string
}

type Memory struct {
	Size  int
	Model string
}

type Disk struct {
	Size int
}

type HostPlugin struct {
	AppInstanceRef AppInstanceRef
	AppRef         AppRef
}

type SdkPlugin struct {
	AppInstanceRef AppInstanceRef
	AppRef         AppRef
}

func (obj Host) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

func (obj *Host) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

func (obj Host) SpecHash() string {
	data, _ := json.Marshal(&obj.Spec.SSH)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

type HostRegistry struct {
	registry.Registry
}

func hostPreCreate(obj core.ApiObject) error {
	host := obj.(*Host)
	host.Metadata.Finalizers = []string{core.FinalizerCleanRefGPU, core.FinalizerCleanRefEvent}
	return nil
}

func NewHost() *Host {
	host := new(Host)
	host.Init(ApiVersion, core.KindHost)
	host.Spec.Plugins = []HostPlugin{}
	host.Spec.Sdks = []SdkPlugin{}
	return host
}

func NewHostRegistry() HostRegistry {
	r := HostRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindHost), false),
	}
	r.SetPreCreateHook(hostPreCreate)
	return r
}
