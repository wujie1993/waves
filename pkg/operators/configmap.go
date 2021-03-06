package operators

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

// ConfigMapOperator 配置字典管理器
type ConfigMapOperator struct {
	BaseOperator

	revisioner registry.Revisioner
}

// handleConfigMap 处理配置字典的变更操作
func (o *ConfigMapOperator) handleConfigMap(ctx context.Context, obj core.ApiObject) error {
	configMap := obj.(*v1.ConfigMap)
	log.Tracef("%s '%s' is %s", configMap.Kind, configMap.GetKey(), configMap.Status.Phase)

	switch configMap.Status.Phase {
	case core.PhaseDeleting:
		o.delete(ctx, obj)
	}
	return nil
}

// finalizeConfigMap 级联清除配置字典的关联资源
func (o ConfigMapOperator) finalizeConfigMap(ctx context.Context, obj core.ApiObject) error {
	configMap := obj.(*v1.ConfigMap)

	// 每次只处理一项Finalizer
	switch configMap.Metadata.Finalizers[0] {
	case core.FinalizerCleanRevision:
		if err := o.revisioner.DeleteAllRevisions(ctx, configMap.Metadata.Namespace, configMap.Metadata.Name); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

// NewConfigMapOperator 创建配置字典管理器
func NewConfigMapOperator() *ConfigMapOperator {
	o := &ConfigMapOperator{
		BaseOperator: NewBaseOperator(v1.NewConfigMapRegistry()),
		revisioner:   v1.NewConfigMapRevision(),
	}
	o.SetHandleFunc(o.handleConfigMap)
	o.SetFinalizeFunc(o.finalizeConfigMap)
	return o
}
