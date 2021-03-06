package v1

import (
	"context"
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/orm/core"
)

// ConfigMapRevision 配置字典修订版本记录器，实现了Revisioner接口
type ConfigMapRevision struct {
	kind string
}

func (r ConfigMapRevision) SetRevision(ctx context.Context, obj core.ApiObject) error {
	configMap := obj.(*ConfigMap)

	// 如果与上个版本无差异，则不再创建新的历史版本
	lastRevision, err := r.GetLastRevision(ctx, configMap.Metadata.Namespace, configMap.Metadata.Name)
	if err != nil {
		return err
	}
	if lastRevision != nil && lastRevision.SpecHash() == obj.SpecHash() {
		return nil
	}

	revision := NewRevision()
	revision.Metadata.Name = fmt.Sprintf("%s-%d-%s", configMap.Metadata.Name, configMap.Metadata.ResourceVersion, configMap.SpecHash())
	revision.ResourceRef = ResourceRef{
		Kind:      core.KindConfigMap,
		Namespace: configMap.Metadata.Namespace,
		Name:      configMap.Metadata.Name,
	}
	revision.Revision = configMap.Metadata.ResourceVersion
	data, err := configMap.ToJSON()
	if err != nil {
		return err
	}
	revision.Data = string(data)

	revisionRegistry := NewRevisionRegistry()
	if _, err := revisionRegistry.Create(context.TODO(), revision); err != nil {
		return err
	}

	return nil
}

func (r ConfigMapRevision) ListRevisions(ctx context.Context, namespace string, name string) (core.ApiObjectList, error) {
	configMapRegistry := NewConfigMapRegistry()

	obj, err := configMapRegistry.Get(ctx, namespace, name)
	if err != nil {
		return nil, err
	} else if obj == nil {
		return nil, nil
	}

	revisionRegistry := NewRevisionRegistry()

	revisionList, err := revisionRegistry.List(context.TODO(), "")
	if err != nil {
		return nil, err
	}

	result := []core.ApiObject{}
	for _, revisionObj := range revisionList {
		revision := revisionObj.(*Revision)
		if revision.ResourceRef.Kind == r.kind && revision.ResourceRef.Namespace == namespace && revision.ResourceRef.Name == name {
			item, err := New(r.kind)
			if err != nil {
				return nil, err
			}
			if err := item.FromJSON([]byte(revision.Data)); err != nil {
				return nil, err
			}

			result = append(result, item)
		}
	}

	sort.Sort(sort.Reverse(core.SortByRevision(result)))

	return result, nil
}

func (r ConfigMapRevision) GetRevision(ctx context.Context, namespace string, name string, revision int) (core.ApiObject, error) {
	configMapRegistry := NewConfigMapRegistry()

	obj, err := configMapRegistry.Get(ctx, namespace, name)
	if err != nil {
		return nil, err
	} else if obj == nil {
		return nil, nil
	}
	configMap := obj.(*ConfigMap)
	if revision >= configMap.Metadata.ResourceVersion {
		return nil, nil
	}

	revisionRegistry := NewRevisionRegistry()

	revisionList, err := revisionRegistry.List(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, revisionObj := range revisionList {
		rev := revisionObj.(*Revision)
		if rev.ResourceRef.Kind == r.kind && rev.ResourceRef.Namespace == namespace && rev.ResourceRef.Name == name && rev.Revision == revision {
			result, err := New(r.kind)
			if err != nil {
				return nil, err
			}
			if err := result.FromJSON([]byte(rev.Data)); err != nil {
				return nil, err
			}

			return result, nil
		}
	}
	return nil, nil
}

func (r ConfigMapRevision) RevertRevision(ctx context.Context, namespace string, name string, revision int) (core.ApiObject, error) {
	configMapRegistry := NewConfigMapRegistry()

	obj, err := r.GetRevision(ctx, namespace, name, revision)
	if err != nil {
		log.Error(err)
		return nil, err
	} else if obj == nil {
		return nil, nil
	}

	return configMapRegistry.Update(ctx, obj)
}

func (r ConfigMapRevision) GetLastRevision(ctx context.Context, namespace string, name string) (core.ApiObject, error) {
	objs, err := r.ListRevisions(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	if len(objs) > 0 {
		return objs[0], nil
	}
	return nil, nil
}

func (r ConfigMapRevision) DeleteRevision(ctx context.Context, namespace string, name string, revision int) (core.ApiObject, error) {
	configMapRegistry := NewConfigMapRegistry()

	obj, err := configMapRegistry.Get(ctx, namespace, name)
	if err != nil {
		return nil, err
	} else if obj == nil {
		return nil, nil
	}

	resourceVersion := obj.GetMetadata().ResourceVersion
	if revision >= resourceVersion || resourceVersion <= 0 {
		return nil, nil
	}

	revisionRegistry := NewRevisionRegistry()

	revisionList, err := revisionRegistry.List(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, revisionObj := range revisionList {
		rev := revisionObj.(*Revision)
		if rev.ResourceRef.Kind == r.kind && rev.ResourceRef.Namespace == namespace && rev.ResourceRef.Name == name && rev.Revision == revision {
			result, err := New(r.kind)
			if err != nil {
				return nil, err
			}
			if err := result.FromJSON([]byte(rev.Data)); err != nil {
				return nil, err
			}

			if _, err := revisionRegistry.Delete(ctx, "", rev.Metadata.Name); err != nil {
				return nil, err
			}

			return result, nil
		}
	}
	return nil, nil
}

func (r ConfigMapRevision) DeleteAllRevisions(ctx context.Context, namespace string, name string) error {
	configMapRegistry := NewConfigMapRegistry()

	obj, err := configMapRegistry.Get(ctx, namespace, name)
	if err != nil {
		return err
	} else if obj == nil {
		return nil
	}

	resourceVersion := obj.GetMetadata().ResourceVersion
	if resourceVersion <= 0 {
		return nil
	}

	revisionRegistry := NewRevisionRegistry()

	revisionList, err := revisionRegistry.List(ctx, "")
	if err != nil {
		return err
	}

	for _, revisionObj := range revisionList {
		rev := revisionObj.(*Revision)
		if rev.ResourceRef.Kind == r.kind && rev.ResourceRef.Namespace == namespace && rev.ResourceRef.Name == name {
			if _, err := revisionRegistry.Delete(ctx, "", rev.Metadata.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

// NewConfigMapRevision 实例化配置字典修订历史记录器
func NewConfigMapRevision() *ConfigMapRevision {
	return &ConfigMapRevision{
		kind: core.KindConfigMap,
	}
}
