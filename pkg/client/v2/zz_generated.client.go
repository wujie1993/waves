// Code generated by codegen. DO NOT EDIT!!!

package v2

import (
	"context"

	"github.com/wujie1993/waves/pkg/client/rest"
	objv2 "github.com/wujie1993/waves/pkg/orm/v2"
)

type Client struct {
	rest.RESTClient
}

func (c Client) AppInstances(namespace string) appinstances {
	return appinstances{
		namespace:  namespace,
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

func NewClient(cli rest.RESTClient) Client {
	return Client{
		RESTClient: cli,
	}
}

type appinstances struct {
	rest.RESTClient
	namespace string
}

func (c appinstances) Get(ctx context.Context, name string) (*objv2.AppInstance, error) {
	result := &objv2.AppInstance{}
	if err := c.RESTClient.Get().
		Version("v2").
		Namespace(c.namespace).
		Resource("appinstances").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c appinstances) Create(ctx context.Context, obj *objv2.AppInstance) (*objv2.AppInstance, error) {
	result := &objv2.AppInstance{}
	if err := c.RESTClient.Post().
		Version("v2").
		Namespace(c.namespace).
		Resource("appinstances").
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c appinstances) List(ctx context.Context) ([]objv2.AppInstance, error) {
	result := []objv2.AppInstance{}
	if err := c.RESTClient.Get().
		Version("v2").
		Namespace(c.namespace).
		Resource("appinstances").
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c appinstances) Update(ctx context.Context, obj *objv2.AppInstance) (*objv2.AppInstance, error) {
	result := &objv2.AppInstance{}
	if err := c.RESTClient.Put().
		Version("v2").
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

func (c appinstances) Delete(ctx context.Context, name string) (*objv2.AppInstance, error) {
	result := &objv2.AppInstance{}
	if err := c.RESTClient.Delete().
		Version("v2").
		Namespace(c.namespace).
		Resource("appinstances").
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

func (c hosts) Get(ctx context.Context, name string) (*objv2.Host, error) {
	result := &objv2.Host{}
	if err := c.RESTClient.Get().
		Version("v2").
		Resource("hosts").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c hosts) Create(ctx context.Context, obj *objv2.Host) (*objv2.Host, error) {
	result := &objv2.Host{}
	if err := c.RESTClient.Post().
		Version("v2").
		Resource("hosts").
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c hosts) List(ctx context.Context) ([]objv2.Host, error) {
	result := []objv2.Host{}
	if err := c.RESTClient.Get().
		Version("v2").
		Resource("hosts").
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c hosts) Update(ctx context.Context, obj *objv2.Host) (*objv2.Host, error) {
	result := &objv2.Host{}
	if err := c.RESTClient.Put().
		Version("v2").
		Resource("hosts").
		Name(obj.Metadata.Name).
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c hosts) Delete(ctx context.Context, name string) (*objv2.Host, error) {
	result := &objv2.Host{}
	if err := c.RESTClient.Delete().
		Version("v2").
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

func (c jobs) Get(ctx context.Context, name string) (*objv2.Job, error) {
	result := &objv2.Job{}
	if err := c.RESTClient.Get().
		Version("v2").
		Resource("jobs").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c jobs) Create(ctx context.Context, obj *objv2.Job) (*objv2.Job, error) {
	result := &objv2.Job{}
	if err := c.RESTClient.Post().
		Version("v2").
		Resource("jobs").
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c jobs) List(ctx context.Context) ([]objv2.Job, error) {
	result := []objv2.Job{}
	if err := c.RESTClient.Get().
		Version("v2").
		Resource("jobs").
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c jobs) Update(ctx context.Context, obj *objv2.Job) (*objv2.Job, error) {
	result := &objv2.Job{}
	if err := c.RESTClient.Put().
		Version("v2").
		Resource("jobs").
		Name(obj.Metadata.Name).
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c jobs) Delete(ctx context.Context, name string) (*objv2.Job, error) {
	result := &objv2.Job{}
	if err := c.RESTClient.Delete().
		Version("v2").
		Resource("jobs").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}
