// Code generated by codegen. DO NOT EDIT!!!

package v1

import (
	"context"

	"github.com/wujie1993/waves/pkg/client/rest"
	objv1 "github.com/wujie1993/waves/pkg/orm/v1"
)

type Client struct {
	rest.RESTClient
}

func (c Client) Apps(namespace string) apps {
	return apps{
		namespace:  namespace,
		RESTClient: c.RESTClient,
	}
}

func (c Client) AppInstances(namespace string) appinstances {
	return appinstances{
		namespace:  namespace,
		RESTClient: c.RESTClient,
	}
}

func (c Client) Audits() audits {
	return audits{
		RESTClient: c.RESTClient,
	}
}

func (c Client) ConfigMaps(namespace string) configmaps {
	return configmaps{
		namespace:  namespace,
		RESTClient: c.RESTClient,
	}
}

func (c Client) Events() events {
	return events{
		RESTClient: c.RESTClient,
	}
}

func (c Client) GPUs() gpus {
	return gpus{
		RESTClient: c.RESTClient,
	}
}

func (c Client) Hosts() hosts {
	return hosts{
		RESTClient: c.RESTClient,
	}
}

func (c Client) Jobs() jobs {
	return jobs{
		RESTClient: c.RESTClient,
	}
}

func (c Client) K8sConfigs(namespace string) k8sconfigs {
	return k8sconfigs{
		namespace:  namespace,
		RESTClient: c.RESTClient,
	}
}

func (c Client) Namespaces() namespaces {
	return namespaces{
		RESTClient: c.RESTClient,
	}
}

func (c Client) Pkgs() pkgs {
	return pkgs{
		RESTClient: c.RESTClient,
	}
}

func (c Client) Projects() projects {
	return projects{
		RESTClient: c.RESTClient,
	}
}

func (c Client) Revisions() revisions {
	return revisions{
		RESTClient: c.RESTClient,
	}
}

func NewClient(cli rest.RESTClient) Client {
	return Client{
		RESTClient: cli,
	}
}

type apps struct {
	rest.RESTClient
	namespace string
}

func (c apps) Get(ctx context.Context, name string) (*objv1.App, error) {
	result := &objv1.App{}
	if err := c.RESTClient.Get().
		Version("v1").
		Namespace(c.namespace).
		Resource("apps").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c apps) Create(ctx context.Context, obj *objv1.App) (*objv1.App, error) {
	result := &objv1.App{}
	if err := c.RESTClient.Post().
		Version("v1").
		Namespace(c.namespace).
		Resource("apps").
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c apps) List(ctx context.Context) ([]objv1.App, error) {
	result := []objv1.App{}
	if err := c.RESTClient.Get().
		Version("v1").
		Namespace(c.namespace).
		Resource("apps").
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c apps) Update(ctx context.Context, obj *objv1.App) (*objv1.App, error) {
	result := &objv1.App{}
	if err := c.RESTClient.Put().
		Version("v1").
		Namespace(c.namespace).
		Resource("apps").
		Name(obj.Metadata.Name).
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c apps) Delete(ctx context.Context, name string) (*objv1.App, error) {
	result := &objv1.App{}
	if err := c.RESTClient.Delete().
		Version("v1").
		Namespace(c.namespace).
		Resource("apps").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

type appinstances struct {
	rest.RESTClient
	namespace string
}

func (c appinstances) Get(ctx context.Context, name string) (*objv1.AppInstance, error) {
	result := &objv1.AppInstance{}
	if err := c.RESTClient.Get().
		Version("v1").
		Namespace(c.namespace).
		Resource("appinstances").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c appinstances) Create(ctx context.Context, obj *objv1.AppInstance) (*objv1.AppInstance, error) {
	result := &objv1.AppInstance{}
	if err := c.RESTClient.Post().
		Version("v1").
		Namespace(c.namespace).
		Resource("appinstances").
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c appinstances) List(ctx context.Context) ([]objv1.AppInstance, error) {
	result := []objv1.AppInstance{}
	if err := c.RESTClient.Get().
		Version("v1").
		Namespace(c.namespace).
		Resource("appinstances").
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c appinstances) Update(ctx context.Context, obj *objv1.AppInstance) (*objv1.AppInstance, error) {
	result := &objv1.AppInstance{}
	if err := c.RESTClient.Put().
		Version("v1").
		Namespace(c.namespace).
		Resource("appinstances").
		Name(obj.Metadata.Name).
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c appinstances) Delete(ctx context.Context, name string) (*objv1.AppInstance, error) {
	result := &objv1.AppInstance{}
	if err := c.RESTClient.Delete().
		Version("v1").
		Namespace(c.namespace).
		Resource("appinstances").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

type audits struct {
	rest.RESTClient
	namespace string
}

func (c audits) Get(ctx context.Context, name string) (*objv1.Audit, error) {
	result := &objv1.Audit{}
	if err := c.RESTClient.Get().
		Version("v1").
		Resource("audits").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c audits) Create(ctx context.Context, obj *objv1.Audit) (*objv1.Audit, error) {
	result := &objv1.Audit{}
	if err := c.RESTClient.Post().
		Version("v1").
		Resource("audits").
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c audits) List(ctx context.Context) ([]objv1.Audit, error) {
	result := []objv1.Audit{}
	if err := c.RESTClient.Get().
		Version("v1").
		Resource("audits").
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c audits) Update(ctx context.Context, obj *objv1.Audit) (*objv1.Audit, error) {
	result := &objv1.Audit{}
	if err := c.RESTClient.Put().
		Version("v1").
		Resource("audits").
		Name(obj.Metadata.Name).
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c audits) Delete(ctx context.Context, name string) (*objv1.Audit, error) {
	result := &objv1.Audit{}
	if err := c.RESTClient.Delete().
		Version("v1").
		Resource("audits").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

type configmaps struct {
	rest.RESTClient
	namespace string
}

func (c configmaps) Get(ctx context.Context, name string) (*objv1.ConfigMap, error) {
	result := &objv1.ConfigMap{}
	if err := c.RESTClient.Get().
		Version("v1").
		Namespace(c.namespace).
		Resource("configmaps").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c configmaps) Create(ctx context.Context, obj *objv1.ConfigMap) (*objv1.ConfigMap, error) {
	result := &objv1.ConfigMap{}
	if err := c.RESTClient.Post().
		Version("v1").
		Namespace(c.namespace).
		Resource("configmaps").
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c configmaps) List(ctx context.Context) ([]objv1.ConfigMap, error) {
	result := []objv1.ConfigMap{}
	if err := c.RESTClient.Get().
		Version("v1").
		Namespace(c.namespace).
		Resource("configmaps").
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c configmaps) Update(ctx context.Context, obj *objv1.ConfigMap) (*objv1.ConfigMap, error) {
	result := &objv1.ConfigMap{}
	if err := c.RESTClient.Put().
		Version("v1").
		Namespace(c.namespace).
		Resource("configmaps").
		Name(obj.Metadata.Name).
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c configmaps) Delete(ctx context.Context, name string) (*objv1.ConfigMap, error) {
	result := &objv1.ConfigMap{}
	if err := c.RESTClient.Delete().
		Version("v1").
		Namespace(c.namespace).
		Resource("configmaps").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

type events struct {
	rest.RESTClient
	namespace string
}

func (c events) Get(ctx context.Context, name string) (*objv1.Event, error) {
	result := &objv1.Event{}
	if err := c.RESTClient.Get().
		Version("v1").
		Resource("events").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c events) Create(ctx context.Context, obj *objv1.Event) (*objv1.Event, error) {
	result := &objv1.Event{}
	if err := c.RESTClient.Post().
		Version("v1").
		Resource("events").
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c events) List(ctx context.Context) ([]objv1.Event, error) {
	result := []objv1.Event{}
	if err := c.RESTClient.Get().
		Version("v1").
		Resource("events").
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c events) Update(ctx context.Context, obj *objv1.Event) (*objv1.Event, error) {
	result := &objv1.Event{}
	if err := c.RESTClient.Put().
		Version("v1").
		Resource("events").
		Name(obj.Metadata.Name).
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c events) Delete(ctx context.Context, name string) (*objv1.Event, error) {
	result := &objv1.Event{}
	if err := c.RESTClient.Delete().
		Version("v1").
		Resource("events").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

type gpus struct {
	rest.RESTClient
	namespace string
}

func (c gpus) Get(ctx context.Context, name string) (*objv1.GPU, error) {
	result := &objv1.GPU{}
	if err := c.RESTClient.Get().
		Version("v1").
		Resource("gpus").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c gpus) Create(ctx context.Context, obj *objv1.GPU) (*objv1.GPU, error) {
	result := &objv1.GPU{}
	if err := c.RESTClient.Post().
		Version("v1").
		Resource("gpus").
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c gpus) List(ctx context.Context) ([]objv1.GPU, error) {
	result := []objv1.GPU{}
	if err := c.RESTClient.Get().
		Version("v1").
		Resource("gpus").
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c gpus) Update(ctx context.Context, obj *objv1.GPU) (*objv1.GPU, error) {
	result := &objv1.GPU{}
	if err := c.RESTClient.Put().
		Version("v1").
		Resource("gpus").
		Name(obj.Metadata.Name).
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c gpus) Delete(ctx context.Context, name string) (*objv1.GPU, error) {
	result := &objv1.GPU{}
	if err := c.RESTClient.Delete().
		Version("v1").
		Resource("gpus").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

type hosts struct {
	rest.RESTClient
	namespace string
}

func (c hosts) Get(ctx context.Context, name string) (*objv1.Host, error) {
	result := &objv1.Host{}
	if err := c.RESTClient.Get().
		Version("v1").
		Resource("hosts").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c hosts) Create(ctx context.Context, obj *objv1.Host) (*objv1.Host, error) {
	result := &objv1.Host{}
	if err := c.RESTClient.Post().
		Version("v1").
		Resource("hosts").
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c hosts) List(ctx context.Context) ([]objv1.Host, error) {
	result := []objv1.Host{}
	if err := c.RESTClient.Get().
		Version("v1").
		Resource("hosts").
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c hosts) Update(ctx context.Context, obj *objv1.Host) (*objv1.Host, error) {
	result := &objv1.Host{}
	if err := c.RESTClient.Put().
		Version("v1").
		Resource("hosts").
		Name(obj.Metadata.Name).
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c hosts) Delete(ctx context.Context, name string) (*objv1.Host, error) {
	result := &objv1.Host{}
	if err := c.RESTClient.Delete().
		Version("v1").
		Resource("hosts").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

type jobs struct {
	rest.RESTClient
	namespace string
}

func (c jobs) Get(ctx context.Context, name string) (*objv1.Job, error) {
	result := &objv1.Job{}
	if err := c.RESTClient.Get().
		Version("v1").
		Resource("jobs").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c jobs) Create(ctx context.Context, obj *objv1.Job) (*objv1.Job, error) {
	result := &objv1.Job{}
	if err := c.RESTClient.Post().
		Version("v1").
		Resource("jobs").
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c jobs) List(ctx context.Context) ([]objv1.Job, error) {
	result := []objv1.Job{}
	if err := c.RESTClient.Get().
		Version("v1").
		Resource("jobs").
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c jobs) Update(ctx context.Context, obj *objv1.Job) (*objv1.Job, error) {
	result := &objv1.Job{}
	if err := c.RESTClient.Put().
		Version("v1").
		Resource("jobs").
		Name(obj.Metadata.Name).
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c jobs) Delete(ctx context.Context, name string) (*objv1.Job, error) {
	result := &objv1.Job{}
	if err := c.RESTClient.Delete().
		Version("v1").
		Resource("jobs").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

type k8sconfigs struct {
	rest.RESTClient
	namespace string
}

func (c k8sconfigs) Get(ctx context.Context, name string) (*objv1.K8sConfig, error) {
	result := &objv1.K8sConfig{}
	if err := c.RESTClient.Get().
		Version("v1").
		Namespace(c.namespace).
		Resource("k8sconfigs").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c k8sconfigs) Create(ctx context.Context, obj *objv1.K8sConfig) (*objv1.K8sConfig, error) {
	result := &objv1.K8sConfig{}
	if err := c.RESTClient.Post().
		Version("v1").
		Namespace(c.namespace).
		Resource("k8sconfigs").
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c k8sconfigs) List(ctx context.Context) ([]objv1.K8sConfig, error) {
	result := []objv1.K8sConfig{}
	if err := c.RESTClient.Get().
		Version("v1").
		Namespace(c.namespace).
		Resource("k8sconfigs").
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c k8sconfigs) Update(ctx context.Context, obj *objv1.K8sConfig) (*objv1.K8sConfig, error) {
	result := &objv1.K8sConfig{}
	if err := c.RESTClient.Put().
		Version("v1").
		Namespace(c.namespace).
		Resource("k8sconfigs").
		Name(obj.Metadata.Name).
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c k8sconfigs) Delete(ctx context.Context, name string) (*objv1.K8sConfig, error) {
	result := &objv1.K8sConfig{}
	if err := c.RESTClient.Delete().
		Version("v1").
		Namespace(c.namespace).
		Resource("k8sconfigs").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

type namespaces struct {
	rest.RESTClient
	namespace string
}

func (c namespaces) Get(ctx context.Context, name string) (*objv1.Namespace, error) {
	result := &objv1.Namespace{}
	if err := c.RESTClient.Get().
		Version("v1").
		Resource("namespaces").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c namespaces) Create(ctx context.Context, obj *objv1.Namespace) (*objv1.Namespace, error) {
	result := &objv1.Namespace{}
	if err := c.RESTClient.Post().
		Version("v1").
		Resource("namespaces").
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c namespaces) List(ctx context.Context) ([]objv1.Namespace, error) {
	result := []objv1.Namespace{}
	if err := c.RESTClient.Get().
		Version("v1").
		Resource("namespaces").
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c namespaces) Update(ctx context.Context, obj *objv1.Namespace) (*objv1.Namespace, error) {
	result := &objv1.Namespace{}
	if err := c.RESTClient.Put().
		Version("v1").
		Resource("namespaces").
		Name(obj.Metadata.Name).
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c namespaces) Delete(ctx context.Context, name string) (*objv1.Namespace, error) {
	result := &objv1.Namespace{}
	if err := c.RESTClient.Delete().
		Version("v1").
		Resource("namespaces").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

type pkgs struct {
	rest.RESTClient
	namespace string
}

func (c pkgs) Get(ctx context.Context, name string) (*objv1.Pkg, error) {
	result := &objv1.Pkg{}
	if err := c.RESTClient.Get().
		Version("v1").
		Resource("pkgs").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c pkgs) Create(ctx context.Context, obj *objv1.Pkg) (*objv1.Pkg, error) {
	result := &objv1.Pkg{}
	if err := c.RESTClient.Post().
		Version("v1").
		Resource("pkgs").
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c pkgs) List(ctx context.Context) ([]objv1.Pkg, error) {
	result := []objv1.Pkg{}
	if err := c.RESTClient.Get().
		Version("v1").
		Resource("pkgs").
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c pkgs) Update(ctx context.Context, obj *objv1.Pkg) (*objv1.Pkg, error) {
	result := &objv1.Pkg{}
	if err := c.RESTClient.Put().
		Version("v1").
		Resource("pkgs").
		Name(obj.Metadata.Name).
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c pkgs) Delete(ctx context.Context, name string) (*objv1.Pkg, error) {
	result := &objv1.Pkg{}
	if err := c.RESTClient.Delete().
		Version("v1").
		Resource("pkgs").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

type projects struct {
	rest.RESTClient
	namespace string
}

func (c projects) Get(ctx context.Context, name string) (*objv1.Project, error) {
	result := &objv1.Project{}
	if err := c.RESTClient.Get().
		Version("v1").
		Resource("projects").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c projects) Create(ctx context.Context, obj *objv1.Project) (*objv1.Project, error) {
	result := &objv1.Project{}
	if err := c.RESTClient.Post().
		Version("v1").
		Resource("projects").
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c projects) List(ctx context.Context) ([]objv1.Project, error) {
	result := []objv1.Project{}
	if err := c.RESTClient.Get().
		Version("v1").
		Resource("projects").
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c projects) Update(ctx context.Context, obj *objv1.Project) (*objv1.Project, error) {
	result := &objv1.Project{}
	if err := c.RESTClient.Put().
		Version("v1").
		Resource("projects").
		Name(obj.Metadata.Name).
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c projects) Delete(ctx context.Context, name string) (*objv1.Project, error) {
	result := &objv1.Project{}
	if err := c.RESTClient.Delete().
		Version("v1").
		Resource("projects").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

type revisions struct {
	rest.RESTClient
	namespace string
}

func (c revisions) Get(ctx context.Context, name string) (*objv1.Revision, error) {
	result := &objv1.Revision{}
	if err := c.RESTClient.Get().
		Version("v1").
		Resource("revisions").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c revisions) Create(ctx context.Context, obj *objv1.Revision) (*objv1.Revision, error) {
	result := &objv1.Revision{}
	if err := c.RESTClient.Post().
		Version("v1").
		Resource("revisions").
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c revisions) List(ctx context.Context) ([]objv1.Revision, error) {
	result := []objv1.Revision{}
	if err := c.RESTClient.Get().
		Version("v1").
		Resource("revisions").
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c revisions) Update(ctx context.Context, obj *objv1.Revision) (*objv1.Revision, error) {
	result := &objv1.Revision{}
	if err := c.RESTClient.Put().
		Version("v1").
		Resource("revisions").
		Name(obj.Metadata.Name).
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c revisions) Delete(ctx context.Context, name string) (*objv1.Revision, error) {
	result := &objv1.Revision{}
	if err := c.RESTClient.Delete().
		Version("v1").
		Resource("revisions").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}
