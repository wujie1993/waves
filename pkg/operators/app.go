package operators

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
)

// AppOperator 应用管理器
type AppOperator struct {
	BaseOperator
}

// handleAppInstance 处理应用实例的变更操作
func (o *AppOperator) handleApp(ctx context.Context, obj core.ApiObject) error {
	app := obj.(*v1.App)
	log.Tracef("%s '%s' is %s", app.Kind, app.GetKey(), app.Status.Phase)

	switch app.Status.Phase {
	case core.PhaseDeleting:
		o.delete(ctx, obj)
	}
	return nil
}

// finalizeApp 级联清除应用的关联资源
func (o AppOperator) finalizeApp(ctx context.Context, obj core.ApiObject) error {
	app := obj.(*v1.App)

	// 每次只处理一项Finalizer
	switch app.Metadata.Finalizers[0] {
	case core.FinalizerCleanRefConfigMap:
		// 同步删除关联的ConfigMap
		for _, versionApp := range app.Spec.Versions {
			for _, module := range versionApp.Modules {
				if module.ConfigMapRef.Name != "" && module.ConfigMapRef.Namespace != "" {
					if _, err := o.helper.V1.ConfigMap.Delete(context.TODO(), module.ConfigMapRef.Namespace, module.ConfigMapRef.Name); err != nil {
						log.Error(err)
						return err
					}
				}
				if module.AdditionalConfigs.ConfigMapRef.Name != "" && module.AdditionalConfigs.ConfigMapRef.Namespace != "" {
					if _, err := o.helper.V1.ConfigMap.Delete(context.TODO(), module.AdditionalConfigs.ConfigMapRef.Namespace, module.AdditionalConfigs.ConfigMapRef.Name); err != nil {
						log.Error(err)
						return err
					}
				}
			}
		}
	}
	return nil
}

// NewAppOperator 创建应用管理器
func NewAppOperator() *AppOperator {
	o := &AppOperator{
		BaseOperator: NewBaseOperator(v1.NewAppRegistry()),
	}
	o.SetFinalizeFunc(o.finalizeApp)
	o.SetHandleFunc(o.handleApp)
	return o
}
