package operators

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/wujie1993/waves/pkg/ansible"
	"github.com/wujie1993/waves/pkg/db"
	"github.com/wujie1993/waves/pkg/e"
	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/v1"
	"github.com/wujie1993/waves/pkg/orm/v2"
	"github.com/wujie1993/waves/pkg/setting"
)

const (
	ArgTypeInteger = "integer"
	ArgTypeNumber  = "number"
	ArgTypeString  = "string"
	ArgTypeBoolean = "boolean"

	ArgFormatInt32     = "int32"
	ArgFormatInt64     = "int64"
	ArgFormatPort      = "port"
	ArgFormatFloat     = "float"
	ArgFormatDouble    = "double"
	ArgFormatDate      = "date"
	ArgFormatPassword  = "password"
	ArgFormatArray     = "array"
	ArgFormatGroupHost = "groupHost"
)

// AppInstanceOperator 应用实例控制器用于处理应用实例的添加，更新和删除行为, 每个应用实例都会与一个应用版本绑定, 根据操作行为的不同, 会创建对应的任务进行处理.
type AppInstanceOperator struct {
	BaseOperator

	// 存放已经激活健康检查（即.Status.Phase为Installed）的应用实例UID与对应的健康检查协程终止方法
	healthCheckMap map[string]context.CancelFunc
	// healthCheckMap锁
	healthCheckMutex sync.Mutex
}

func (o *AppInstanceOperator) handleAppInstance(ctx context.Context, obj core.ApiObject) {
	appInstance := obj.(*v2.AppInstance)
	log.Infof("%s '%s' is %s", appInstance.Kind, appInstance.GetKey(), appInstance.Status.Phase)

	/*
		if appInstance.Status.Phase != core.PhaseWaiting && appInstance.Status.Phase != core.PhaseDeleting {
			// 避免重复更新
			if hash, ok := o.applyings.Get(appInstance.GetKey()); ok && hash != appInstance.SpecHash() {
				if _, err := o.helper.V2.AppInstance.UpdateStatusPhase(appInstance.Metadata.Namespace, appInstance.Metadata.Name, core.PhaseWaiting); err != nil {
					log.Error(err)
				}
				return
			}
			if appInstance.Spec.Action != "" {
				if _, err := o.helper.V2.AppInstance.UpdateStatusPhase(appInstance.Metadata.Namespace, appInstance.Metadata.Name, core.PhaseWaiting); err != nil {
					log.Error(err)
				}
				return
			}
		}
	*/

	o.setHealthCheck(ctx, obj)

	// 根据应用实例的状态做对应的处理
	switch appInstance.Status.Phase {
	case core.PhaseWaiting:
		// 处于等待中状态, 根据应用实例的操作行为创建对应的任务, 并绑定到应用实例上

		// 忽略内容体没有发生更新的应用实例
		if hash, ok := o.applyings.Get(appInstance.GetKey()); ok && hash == appInstance.SpecHash() {
			return
		}

		// 填充事件信息与应用实例状态，忽略没有赋予合法操作行为的应用实例
		var action string
		switch appInstance.Spec.Action {
		case core.AppActionInstall:
			appInstance.Status.Phase = core.PhaseInstalling
			action = core.EventActionInstall
		case core.AppActionUninstall:
			appInstance.Status.Phase = core.PhaseUninstalling
			action = core.EventActionUninstall
		case core.AppActionConfigure:
			appInstance.Status.Phase = core.PhaseConfiguring
			action = core.EventActionConfigure
		case core.AppActionHealthcheck:
			action = core.EventActionHealthCheck
		case core.AppActionUpgrade:
			// 记录应用实例的内容哈希值, 用于后续比较内容体是否有更新，计算哈希值时忽略Action字段
			appInstance.Spec.Action = ""
			o.applyings.Set(appInstance.GetKey(), appInstance.SpecHash())

			appInstance.Status.Phase = core.PhaseUpgradeing
			if _, err := o.helper.V2.AppInstance.Update(context.TODO(), appInstance, core.WithStatus()); err != nil {
				log.Error(err)
				return
			}
			return
		default:
			return
		}

		// 记录应用实例的内容哈希值, 用于后续比较内容体是否有更新，计算哈希值时忽略Action字段
		appInstance.Spec.Action = ""
		o.applyings.Set(appInstance.GetKey(), appInstance.SpecHash())

		// 根据操作行为构建相应的任务
		jobObj, err := o.setupJob(appInstance, action)
		if err != nil {
			log.Errorf("setup %s job failed of %s: %s", action, appInstance.GetKey(), err)
			o.failback(appInstance, action, err.Error(), nil)
			return
		}
		job := jobObj.(*v2.Job)

		// 将应用实例与任务关联并更新应用实例状态
		appInstance.Metadata.Annotations[core.AnnotationJobPrefix+action] = job.Metadata.Name
		if _, err := o.helper.V2.AppInstance.Update(context.TODO(), appInstance, core.WithStatus()); err != nil {
			log.Error(err)
			o.failback(appInstance, action, err.Error(), job)
			return
		}

		// 记录事件开始
		if err := o.recordEvent(Event{
			BaseApiObj: appInstance.BaseApiObj,
			Action:     action,
			Msg:        "",
			JobRef:     job.Metadata.Name,
			Phase:      core.PhaseWaiting,
		}); err != nil {
			log.Error(err)
		}
	case core.PhaseUninstalling:
		// 如果应用实例没有绑定卸载任务，则将应用状态重置为等待中
		jobName, ok := appInstance.Metadata.Annotations[core.AnnotationJobPrefix+core.EventActionUninstall]
		if !ok {
			if _, err := o.helper.V2.AppInstance.UpdateStatusPhase(appInstance.Metadata.Namespace, appInstance.Metadata.Name, core.PhaseWaiting); err != nil {
				log.Error(err)
			}
			return
		}
		// 移除任务注解，需要在后续逻辑中更新应用实例
		delete(appInstance.Metadata.Annotations, core.AnnotationJobPrefix+core.EventActionUninstall)

		// 侦听卸载任务的状态，并在任务执行完成时将应用实例状态置为已卸载
		o.watchAndHandleJob(ctx, jobName, func(job *v2.Job) bool {
			switch job.Status.Phase {
			case core.PhaseCompleted:
				// 释放算法实例GPU
				if err := o.releaseGPU(appInstance); err != nil {
					log.Error(err)
					o.failback(appInstance, core.EventActionUninstall, err.Error(), job)
					return true
				}

				// 如果初始化任务执行成功, 将应用实例状态更新为已卸载并结束任务侦听
				appInstance.Status.SetCondition(core.ConditionTypeInstalled, core.ConditionStatusFalse)
				appInstance.Status.UnsetCondition(core.ConditionTypeHealthy)
				appInstance.SetStatusPhase(core.PhaseUninstalled)
				if _, err := o.helper.V2.AppInstance.Update(context.TODO(), appInstance, core.WithStatus()); err != nil {
					log.Error(err)
					return true
				}

				// 记录事件完成
				if err := o.recordEvent(Event{
					BaseApiObj: appInstance.BaseApiObj,
					Action:     core.EventActionUninstall,
					Msg:        "",
					JobRef:     jobName,
					Phase:      core.PhaseCompleted,
				}); err != nil {
					log.Error(err)
				}
				return true
			case core.PhaseFailed:
				o.failback(appInstance, core.EventActionUninstall, "", job)
				return true
			case core.PhaseWaiting, core.PhaseRunning:
				// 处于运行中状态不做任何处理
				return false
			default:
				log.Warnf("unknown status phase '%s' of job '%s'", job.Status.Phase, jobName)
				return false
			}
		})
	case core.PhaseConfiguring:
		// 如果应用实例没有绑定配置任务，则将应用状态重置为等待中
		jobName, ok := appInstance.Metadata.Annotations[core.AnnotationJobPrefix+core.EventActionConfigure]
		if !ok {
			if _, err := o.helper.V2.AppInstance.UpdateStatusPhase(appInstance.Metadata.Namespace, appInstance.Metadata.Name, core.PhaseWaiting); err != nil {
				log.Error(err)
			}
			return
		}
		// 移除配置任务注解，需要在后续逻辑中更新应用实例
		delete(appInstance.Metadata.Annotations, core.AnnotationJobPrefix+core.EventActionConfigure)

		// 侦听配置任务的状态，并在任务执行完成时将应用实例状态置为已安装
		o.watchAndHandleJob(ctx, jobName, func(job *v2.Job) bool {
			switch job.Status.Phase {
			case core.PhaseWaiting, core.PhaseRunning:
				// 任务运行中，不做任何处理
				return false
			case core.PhaseCompleted:
				// 清除关联的自定义配置
				for moduleIndex, module := range appInstance.Spec.Modules {
					for replicaIndex, replica := range module.Replicas {
						if replica.AdditionalConfigMapRef.Name != "" {
							appInstance.Spec.Modules[moduleIndex].Replicas[replicaIndex].AdditionalConfigMapRef.Name = ""
							configMapDeleteCtx, _ := context.WithTimeout(ctx, time.Second*5)
							if _, err := o.helper.V1.ConfigMap.Delete(configMapDeleteCtx, replica.AdditionalConfigMapRef.Namespace, replica.AdditionalConfigMapRef.Name, core.WithSync()); err != nil {
								log.Error(err)
							}
						}
					}
				}

				appInstance.Status.SetCondition(core.ConditionTypeConfigured, core.ConditionStatusTrue)
				appInstance.SetStatusPhase(core.PhaseInstalled)
				if _, err := o.helper.V2.AppInstance.Update(context.TODO(), appInstance, core.WithStatus()); err != nil {
					log.Error(err)
					return true
				}

				// 记录事件完成
				if err := o.recordEvent(Event{
					BaseApiObj: appInstance.BaseApiObj,
					Action:     core.EventActionConfigure,
					Msg:        "",
					JobRef:     jobName,
					Phase:      core.PhaseCompleted,
				}); err != nil {
					log.Error(err)
				}
				return true
			case core.PhaseFailed:
				o.failback(appInstance, core.EventActionConfigure, "", job)
				return true
			default:
				log.Warnf("unknown status phase '%s' of job '%s'", job.Status.Phase, jobName)
				return false
			}
		})
	case core.PhaseInstalling:
		// 如果应用实例没有绑定安装任务，则将应用状态重置为等待中
		jobName, ok := appInstance.Metadata.Annotations[core.AnnotationJobPrefix+core.EventActionInstall]
		if !ok {
			if _, err := o.helper.V2.AppInstance.UpdateStatusPhase(appInstance.Metadata.Namespace, appInstance.Metadata.Name, core.PhaseWaiting); err != nil {
				log.Error(err)
			}
			return
		}
		// 移除安装任务注解，需要在后续逻辑中更新应用实例
		delete(appInstance.Metadata.Annotations, core.AnnotationJobPrefix+core.EventActionInstall)

		// 侦听安装任务的状态，并在任务执行完成时将应用实例状态置为已就绪
		o.watchAndHandleJob(ctx, jobName, func(job *v2.Job) bool {
			switch job.Status.Phase {
			case core.PhaseWaiting, core.PhaseRunning:
				// 任务运行中，不做任何处理
				return false
			case core.PhaseCompleted:
				// 如果任务执行成功, 将应用实例状态更新为已安装并结束任务侦听
				appInstance.Status.SetCondition(core.ConditionTypeInstalled, core.ConditionStatusTrue)
				appInstance.SetStatusPhase(core.PhaseInstalled)
				if _, err := o.helper.V2.AppInstance.Update(context.TODO(), appInstance, core.WithStatus()); err != nil {
					log.Error(err)
					return true
				}

				// 记录事件完成
				if err := o.recordEvent(Event{
					BaseApiObj: appInstance.BaseApiObj,
					Action:     core.EventActionInstall,
					Msg:        "",
					JobRef:     jobName,
					Phase:      core.PhaseCompleted,
				}); err != nil {
					log.Error(err)
				}
				return true
			case core.PhaseFailed:
				o.failback(appInstance, core.EventActionInstall, "", job)
				return true
			default:
				log.Warnf("unknown status phase '%s' of job '%s'", job.Status.Phase, jobName)
				return false
			}
		})
	case core.PhaseUpgradeing:
		o.upgradeAppInstance(ctx, obj)
	case core.PhaseUpgradeBackoffing:
		o.upgradeBackoffAppInstance(ctx, obj)
	case core.PhaseDeleting:
		// 如果资源正在删除中，则跳过
		if _, ok := o.deletings.Get(appInstance.GetKey()); ok {
			return
		}
		o.deletings.Set(appInstance.GetKey(), appInstance.SpecHash())
		defer o.deletings.Unset(appInstance.GetKey())

		if len(appInstance.Metadata.Finalizers) > 0 {
			// 每次只处理一项Finalizer
			switch appInstance.Metadata.Finalizers[0] {
			case core.FinalizerCleanRefEvent:
				// 同步删除关联的事件
				eventList, err := o.helper.V1.Event.List(context.TODO(), "")
				if err != nil {
					log.Error(err)
					return
				}
				for _, eventObj := range eventList {
					event := eventObj.(*v1.Event)
					if event.Spec.ResourceRef.Kind == core.KindAppInstance && event.Spec.ResourceRef.Namespace == appInstance.Metadata.Namespace && event.Spec.ResourceRef.Name == appInstance.Metadata.Name {
						if _, err := o.helper.V1.Event.Delete(context.TODO(), "", event.Metadata.Name, core.WithSync()); err != nil {
							log.Error(err)
							return
						}
					}
				}
			case core.FinalizerCleanRefConfigMap:
				// 清除关联的配置字典
				for _, module := range appInstance.Spec.Modules {
					for _, replica := range module.Replicas {
						if replica.ConfigMapRef.Name != "" {
							configMapDeleteCtx, _ := context.WithTimeout(ctx, time.Second*5)
							if _, err := o.helper.V1.ConfigMap.Delete(configMapDeleteCtx, replica.ConfigMapRef.Namespace, replica.ConfigMapRef.Name, core.WithSync()); err != nil {
								log.Error(err)
								return
							}
						}
					}
				}
				if appInstance.Spec.Global.ConfigMapRef.Name != "" {
					if _, err := o.helper.V1.ConfigMap.Delete(context.TODO(), appInstance.Spec.Global.ConfigMapRef.Namespace, appInstance.Spec.Global.ConfigMapRef.Name, core.WithSync()); err != nil {
						log.Error(err)
						return
					}
				}
			case core.FinalizerReleaseRefGPU:
				// 释放关联的GPU资源
				if err := o.releaseGPU(appInstance); err != nil {
					log.Error(err)
					return
				}
			}

			o.deletings.Unset(appInstance.GetKey())
			appInstance.Metadata.Finalizers = appInstance.Metadata.Finalizers[1:]
			if _, err := o.helper.V2.AppInstance.Update(context.TODO(), appInstance, core.WithFinalizer()); err != nil {
				log.Error(err)
				return
			}
		} else {
			if _, err := o.helper.V2.AppInstance.Delete(context.TODO(), appInstance.Metadata.Namespace, appInstance.Metadata.Name); err != nil {
				log.Error(err)
				return
			}
		}
	}
}

// 对于已安装状态的应用实例，当应用支持健康检查时，开启健康检查，在其他状态下关闭健康检查
func (o *AppInstanceOperator) setHealthCheck(ctx context.Context, obj core.ApiObject) {
	appInstance := obj.(*v2.AppInstance)

	if appInstance.Status.Phase == core.PhaseInstalled {
		appObj, err := o.helper.V1.App.Get(context.TODO(), core.DefaultNamespace, appInstance.Spec.AppRef.Name)
		if err != nil {
			log.Error(err)
			return
		} else if appObj == nil {
			return
		}

		app := appObj.(*v1.App)
		for _, versionApp := range app.Spec.Versions {
			if versionApp.Version == appInstance.Spec.AppRef.Version {
				for _, supportAction := range versionApp.SupportActions {
					if supportAction == core.AppActionHealthcheck {
						o.enableHealthCheck(ctx, obj)
					}
				}
				return
			}
		}
	} else {
		o.disableHealthCheck(obj)
	}
}

// 侦听任务的变更，并将任务的.Status.Phase变化交由handleJob处理，handleJob返回的bool值表示是否终止任务的侦听
func (o *AppInstanceOperator) watchAndHandleJob(ctx context.Context, jobName string, handleJob func(*v2.Job) bool) error {
	jobCtx, jobWatchCancel := context.WithCancel(ctx)
	defer jobWatchCancel()

	jobActionChan := o.helper.V2.Job.GetWatch(jobCtx, "", jobName)
	for jobAction := range jobActionChan {
		if jobAction.Obj == nil {
			err := e.Errorf("received nil object of job %s", jobName)
			log.Error(err)
			return err
		}
		switch jobAction.Type {
		case db.KVActionTypeDelete:
			err := e.Errorf("job %s has been deleted", jobName)
			log.Error(err)
			return err
		case db.KVActionTypeSet:
			if done := handleJob(jobAction.Obj.(*v2.Job)); done {
				return nil
			}
		}
	}
	return nil
}

// 开启健康检查
func (o *AppInstanceOperator) enableHealthCheck(ctx context.Context, obj core.ApiObject) {
	appInstance := obj.(*v2.AppInstance)

	o.healthCheckMutex.Lock()
	defer o.healthCheckMutex.Unlock()

	_, ok := o.healthCheckMap[appInstance.Metadata.Uid]

	// 如果健康检查已经存在，则跳过
	if ok {
		return
	}

	// 创建健康检查任务
	healthCheckCtx, healthCheckCancel := context.WithCancel(ctx)
	o.healthCheckMap[appInstance.Metadata.Uid] = healthCheckCancel

	go o.healthCheck(healthCheckCtx, appInstance)
}

// 关闭健康检查
func (o *AppInstanceOperator) disableHealthCheck(obj core.ApiObject) {
	appInstance := obj.(*v2.AppInstance)

	o.healthCheckMutex.Lock()
	defer o.healthCheckMutex.Unlock()

	cancel, ok := o.healthCheckMap[appInstance.Metadata.Uid]

	// 如果健康检查已经存在，则取消健康检查，并移除健康状态
	if ok {
		log.Debugf("cancel health check for app instance %s/%s", appInstance.Metadata.Namespace, appInstance.Metadata.Name)
		cancel()
		delete(o.healthCheckMap, appInstance.Metadata.Uid)
		appInstance.Status.UnsetCondition(core.ConditionTypeHealthy)
	}
}

// 执行健康检查
func (o *AppInstanceOperator) healthCheck(ctx context.Context, obj core.ApiObject) {
	appInstance := obj.(*v2.AppInstance)
	// 退出时终止健康检查
	defer func() {
		o.healthCheckMutex.Lock()
		delete(o.healthCheckMap, appInstance.Metadata.Uid)
		o.healthCheckMutex.Unlock()
	}()

	// 首次健康检查延迟
	time.Sleep(time.Duration(appInstance.Spec.LivenessProbe.InitialDelaySeconds) * time.Second)
	select {
	case <-ctx.Done():
		return
	default:
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// 创建健康检查任务
				jobObj, err := o.setupJob(appInstance, core.EventActionHealthCheck)
				if err != nil {
					log.Error(err)
					o.failback(appInstance, core.EventActionHealthCheck, err.Error(), nil)
					return
				}
				job := jobObj.(*v2.Job)

				log.Debugf("running healthcheck of %s", appInstance.GetKey())

				// 侦听健康检查任务，出于性能方面考虑，此处不使用goroutine异步执行，即下一次健康检查的间隔计时是在上一次健康检查结束后才开始
				if err := o.watchAndHandleJob(ctx, job.Metadata.Name, func(job *v2.Job) bool {
					switch job.Status.Phase {
					case core.PhaseWaiting, core.PhaseRunning:
						// 任务运行中，不做任何处理
						return false
					case core.PhaseCompleted:
						log.Debugf("healthcheck succeed of %s", appInstance.GetKey())

						// 清除健康检查历史事件日志
						eventObjs, err := o.helper.V1.Event.List(context.TODO(), "")
						if err != nil {
							log.Error(err)
						} else {
							for _, eventObj := range eventObjs {
								event := eventObj.(*v1.Event)
								if event.Spec.ResourceRef.Namespace == appInstance.Metadata.Namespace && event.Spec.ResourceRef.Name == appInstance.Metadata.Name && event.Spec.ResourceRef.Kind == core.KindAppInstance && event.Spec.Action == core.EventActionHealthCheck {
									if _, err := o.helper.V1.Event.Delete(context.TODO(), "", event.Metadata.Name); err != nil {
										log.Error(err)
									}
								}
							}
						}

						// 如果原来处于非健康状态, 则更新为健康状态
						if appInstance.Status.GetCondition(core.ConditionTypeHealthy) != core.ConditionStatusTrue {
							appInstance.Status.SetCondition(core.ConditionTypeHealthy, core.ConditionStatusTrue)
							if _, err := o.helper.V2.AppInstance.UpdateStatus(appInstance.Metadata.Namespace, appInstance.Metadata.Name, appInstance.Status); err != nil {
								log.Error(err)
							}
						}

						// 记录事件完成
						if err := o.recordEvent(Event{
							BaseApiObj: appInstance.BaseApiObj,
							Action:     core.EventActionHealthCheck,
							Msg:        "",
							JobRef:     job.Metadata.Name,
							Phase:      core.PhaseCompleted,
						}); err != nil {
							log.Error(err)
						}
						return true
					case core.PhaseFailed:
						log.Warnf("healthcheck failed of %s", appInstance.GetKey())
						// 如果任务执行失败，将应用实例置为非健康状态
						o.failback(appInstance, core.EventActionHealthCheck, "", job)
						return true
					default:
						log.Warnf("unknown status phase '%s' of job '%s'", job.Status.Phase, job.GetKey())
						return false
					}
				}); err != nil {
					log.Error(err)
					o.failback(appInstance, core.EventActionHealthCheck, "", job)
					break
				}
			}

			// 健康检查间隔
			time.Sleep(time.Duration(appInstance.Spec.LivenessProbe.PeriodSeconds) * time.Second)
		}
	}
}

// 操作失败回退
func (o AppInstanceOperator) failback(obj core.ApiObject, action string, reason string, job *v2.Job) {
	appInstance := obj.(*v2.AppInstance)

	var jobRef string
	if job != nil {
		jobRef = job.Metadata.Name
		if reason == "" {
			reason = job.Status.GetCondition(core.ConditionTypeRun)
		}
	}

	switch action {
	case core.EventActionInitial:
		appInstance.Status.UnsetCondition(core.ConditionTypeInstalled)
		appInstance.Status.UnsetCondition(core.ConditionTypeHealthy)
	case core.EventActionConfigure:
		appInstance.Status.UnsetCondition(core.ConditionTypeHealthy)
		appInstance.Status.SetCondition(core.ConditionTypeConfigured, reason)
		appInstance.SetStatusPhase(core.PhaseInstalled)
		if _, err := o.helper.V2.AppInstance.Update(context.TODO(), appInstance, core.WithStatus()); err != nil {
			log.Error(err)
		}
		// 记录失败事件
		if err := o.recordEvent(Event{
			BaseApiObj: appInstance.BaseApiObj,
			Action:     action,
			Msg:        reason,
			JobRef:     jobRef,
			Phase:      core.PhaseFailed,
		}); err != nil {
			log.Error(err)
		}
		return
	case core.EventActionInstall:
		appInstance.Status.SetCondition(core.ConditionTypeInstalled, reason)
		appInstance.Status.UnsetCondition(core.ConditionTypeHealthy)
	case core.EventActionUninstall:
		appInstance.Status.SetCondition(core.ConditionTypeInstalled, reason)
		appInstance.Status.UnsetCondition(core.ConditionTypeHealthy)
	case core.EventActionHealthCheck:
		// 清除健康检查历史事件日志
		eventObjs, err := o.helper.V1.Event.List(context.TODO(), "")
		if err != nil {
			log.Error(err)
		} else {
			for _, eventObj := range eventObjs {
				event := eventObj.(*v1.Event)
				if event.Spec.ResourceRef.Namespace == appInstance.Metadata.Namespace && event.Spec.ResourceRef.Name == appInstance.Metadata.Name && event.Spec.ResourceRef.Kind == core.KindAppInstance && event.Spec.Action == core.EventActionHealthCheck {
					if _, err := o.helper.V1.Event.Delete(context.TODO(), "", event.Metadata.Name); err != nil {
						log.Error(err)
					}
				}
			}
		}

		// 记录失败事件
		if err := o.recordEvent(Event{
			BaseApiObj: appInstance.BaseApiObj,
			Action:     action,
			Msg:        reason,
			JobRef:     jobRef,
			Phase:      core.PhaseFailed,
		}); err != nil {
			log.Error(err)
		}

		// 在健康状态发生变化时更新
		if appInstance.Status.GetCondition(core.ConditionTypeHealthy) != reason {
			appInstance.Status.SetCondition(core.ConditionTypeHealthy, reason)
			if _, err := o.helper.V2.AppInstance.Update(context.TODO(), appInstance, core.WithStatus()); err != nil {
				log.Error(err)
			}
		}
		return
	case core.EventActionUpgrade:
		appInstance.Status.SetCondition(core.ConditionTypeInstalled, reason)
		appInstance.Status.UnsetCondition(core.ConditionTypeHealthy)
		appInstance.SetStatusPhase(core.PhaseUpgradeBackoffing)
		if _, err := o.helper.V2.AppInstance.Update(context.TODO(), appInstance, core.WithStatus()); err != nil {
			log.Error(err)
		}
		// 记录失败事件
		if err := o.recordEvent(Event{
			BaseApiObj: appInstance.BaseApiObj,
			Action:     action,
			Msg:        reason,
			JobRef:     jobRef,
			Phase:      core.PhaseFailed,
		}); err != nil {
			log.Error(err)
		}
		return

	}

	// 更新应用实例状态
	appInstance.Spec.Action = ""
	appInstance.SetStatusPhase(core.PhaseFailed)
	if _, err := o.helper.V2.AppInstance.Update(context.TODO(), appInstance, core.WithStatus()); err != nil {
		log.Error(err)
	}

	// 记录失败事件
	if err := o.recordEvent(Event{
		BaseApiObj: appInstance.BaseApiObj,
		Action:     action,
		Msg:        reason,
		JobRef:     jobRef,
		Phase:      core.PhaseFailed,
	}); err != nil {
		log.Error(err)
	}
}

// 根据操作行为构建任务
func (o *AppInstanceOperator) setupJob(obj core.ApiObject, action string) (core.ApiObject, error) {
	appInstance := obj.(*v2.AppInstance)

	// 获取应用实例对应的应用
	appObj, err := o.helper.V1.App.Get(context.TODO(), o.namespace, appInstance.Spec.AppRef.Name)
	if err != nil {
		err = e.Errorf("failed to get referred app %s: %s", appInstance.Spec.AppRef.Name, err)
		log.Error(err)
		return nil, err
	} else if obj == nil {
		err := e.Errorf("referred app %s not found", appInstance.Spec.AppRef.Name)
		log.Error(err)
		return nil, err
	}
	app := appObj.(*v1.App)

	// 检索应用实例对应的应用版本
	var versionApp *v1.AppVersion
	for _, version := range app.Spec.Versions {
		if version.Version == appInstance.Spec.AppRef.Version {
			if !version.Enabled {
				err := e.Errorf("version %s has been disabled, please make sure package %s still exist", version.Version, version.PkgRef)
				log.Error(err)
				return nil, err
			}
			versionApp = &version
			break
		}
	}
	if versionApp == nil {
		err := e.Errorf("referred app %s does not contain version of %s", appInstance.Spec.AppRef.Name, appInstance.Spec.AppRef.Version)
		log.Error(err)
		return nil, err
	}

	// 生成额外参数
	extraGlobalVars := make(map[string]interface{})
	switch app.Spec.Category {
	case core.AppCategoryCustomize:
		// 填充普通应用和算法插件额外参数
		package_dir, _ := filepath.Abs(filepath.Join(setting.PackageSetting.PkgPath, versionApp.PkgRef))
		extraGlobalVars["package_dir"] = package_dir
	}
	extraGlobalVars["app_name"] = app.Metadata.Name
	extraGlobalVars["deployer_data_dir"], _ = filepath.Abs(setting.AppSetting.DataDir)
	extraGlobalVars["app_instance_id"] = appInstance.Metadata.Uid

	// 构建inventory, playbook与配置文件
	plays := []v2.JobAnsiblePlay{}

	// 生成公共inventory内容
	commonInventoryStr, err := ansible.RenderCommonInventory()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	switch versionApp.Platform {
	case core.AppPlatformBareMetal:
		// 对于部署于裸机平台上的应用实例，以每个子模块为名构建group，子模块参数作为组参数，并将全局参数追加到每个组的参数
		for moduleIndex, module := range appInstance.Spec.Modules {
			for replicaIndex, _ := range module.Replicas {
				play, err := o.setupBareMetalJobPlay(appInstance, moduleIndex, replicaIndex, action, commonInventoryStr, extraGlobalVars, *app)
				if err != nil {
					log.Error(err)
					return nil, err
				}

				plays = append(plays, play)
			}
		}
	case core.AppPlatformK8s:
		// 对于部署于k8s平台上的应用实例，所有的模块都使用[k8s-master]

		// 获取k8s master节点
		host, err := o.helper.V1.K8sConfig.GetFirstMasterHost(appInstance.Spec.K8sRef)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		for moduleIndex, module := range appInstance.Spec.Modules {
			for replicaIndex, _ := range module.Replicas {
				play, err := o.setupK8sJobPlay(appInstance, moduleIndex, replicaIndex, action, commonInventoryStr, extraGlobalVars, *app, host)
				if err != nil {
					log.Error(err)
					return nil, err
				}

				plays = append(plays, play)
			}
		}
	}

	// 创建任务
	job := v2.NewJob()
	job.Metadata.Namespace = o.namespace
	job.Metadata.Name = fmt.Sprintf("%s-%s-%s-%d", core.KindAppInstance, appInstance.Metadata.Name, appInstance.Spec.Action, time.Now().Unix())
	job.Spec.Exec.Type = core.JobExecTypeAnsible
	job.Spec.Exec.Ansible.Bin = setting.AnsibleSetting.Bin
	job.Spec.Exec.Ansible.Plays = plays
	if action == core.EventActionHealthCheck {
		job.Spec.Exec.Ansible.RecklessMode = true
	}
	if action == core.EventActionHealthCheck && appInstance.Spec.LivenessProbe.TimeoutSeconds > 0 {
		job.Spec.TimeoutSeconds = time.Duration(appInstance.Spec.LivenessProbe.TimeoutSeconds)
	} else {
		job.Spec.TimeoutSeconds = 3600
	}
	job.Spec.FailureThreshold = 1
	if _, err := o.helper.V2.Job.Create(context.TODO(), job); err != nil {
		log.Error(err)
		return nil, err
	}

	return job, nil
}

// 构建应用实例的应用版本升级任务，表示将oldAppInstance升级/回退到newAppInstance
func (o AppInstanceOperator) setupUpgradeJob(oldAppInstance *v2.AppInstance, newAppInstance *v2.AppInstance) (*v2.Job, error) {
	// 获取新旧两版应用
	appObj, err := o.helper.V1.App.Get(context.TODO(), core.DefaultNamespace, newAppInstance.Spec.AppRef.Name)
	if err != nil {
		log.Error(err)
		return nil, err
	} else if appObj == nil {
		err := e.Errorf("app %s not found", newAppInstance.Spec.AppRef)
		log.Error(err)
		return nil, err
	}
	app := appObj.(*v1.App)
	var newVersionApp, oldVersionApp v1.AppVersion
	for _, versionApp := range app.Spec.Versions {
		if versionApp.Version == newAppInstance.Spec.AppRef.Version {
			if !versionApp.Enabled {
				err := e.Errorf("The new version %s of the %s app is disabled", newAppInstance.Spec.AppRef.Version, newAppInstance.Spec.AppRef.Name)
				log.Error(err)
				return nil, err
			}
			newVersionApp = versionApp
		}
		if versionApp.Version == oldAppInstance.Spec.AppRef.Version {
			if !versionApp.Enabled {
				err := e.Errorf("The old version %s of the %s app is disabled", oldAppInstance.Spec.AppRef.Version, oldAppInstance.Spec.AppRef.Name)
				log.Error(err)
				return nil, err
			}
			oldVersionApp = versionApp
		}
	}
	if newVersionApp.Version == "" {
		err := e.Errorf("The new version %s of the %s app not found", newAppInstance.Spec.AppRef.Version, newAppInstance.Spec.AppRef.Name)
		log.Error(err)
		return nil, err
	}
	if oldVersionApp.Version == "" {
		err := e.Errorf("The old version %s of the %s app not found", oldAppInstance.Spec.AppRef.Version, oldAppInstance.Spec.AppRef.Name)
		log.Error(err)
		return nil, err
	}

	// 生成额外参数
	extraGlobalVars := make(map[string]interface{})
	extraGlobalVars["app_name"] = app.Metadata.Name
	extraGlobalVars["deployer_data_dir"], _ = filepath.Abs(setting.AppSetting.DataDir)
	extraGlobalVars["app_instance_id"] = newAppInstance.Metadata.Uid

	// 构建inventory, playbook与配置文件
	plays := []v2.JobAnsiblePlay{}

	// 生成公共inventory内容
	commonInventoryStr, err := ansible.RenderCommonInventory()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// 生成新旧版本间应用升级各模块所需执行的操作，根据应用实例的模块版本区分
	oldModuleActions := make(map[string]string)
	for _, oldModule := range oldAppInstance.Spec.Modules {
		newModule, ok := newAppInstance.GetModule(oldModule.Name)
		// 当模块在新版本中不存在，或者存在但版本不同时，卸载旧模块
		if !ok || (ok && oldModule.AppVersion != newModule.AppVersion) {
			oldModuleActions[oldModule.Name] = core.EventActionUninstall
		}
	}
	newModuleActions := make(map[string]string)
	for _, newModule := range newAppInstance.Spec.Modules {
		oldModule, ok := oldAppInstance.GetModule(newModule.Name)
		// 模块在旧版本中存在且版本相同，则更新模块配置，否则安装模块
		if ok && newModule.AppVersion == oldModule.AppVersion {
			newModuleActions[newModule.Name] = core.EventActionConfigure
		} else {
			newModuleActions[newModule.Name] = core.EventActionInstall
		}
	}

	// 填充任务Play
	switch oldVersionApp.Platform {
	case core.AppPlatformBareMetal:
		// 生成旧版本卸载Play
		extraGlobalVars["package_dir"], _ = filepath.Abs(filepath.Join(setting.PackageSetting.PkgPath, oldVersionApp.PkgRef))
		for moduleIndex, module := range oldAppInstance.Spec.Modules {
			oldModuleAction, ok := oldModuleActions[module.Name]
			if !ok {
				continue
			}
			for replicaIndex, _ := range module.Replicas {
				play, err := o.setupBareMetalJobPlay(oldAppInstance, moduleIndex, replicaIndex, oldModuleAction, commonInventoryStr, extraGlobalVars, *app)
				if err != nil {
					log.Error(err)
					return nil, err
				}

				plays = append(plays, play)
			}
		}

		// 生成新版本安装Play
		extraGlobalVars["package_dir"], _ = filepath.Abs(filepath.Join(setting.PackageSetting.PkgPath, newVersionApp.PkgRef))
		for moduleIndex, module := range newAppInstance.Spec.Modules {
			newModuleAction, ok := newModuleActions[module.Name]
			if !ok {
				continue
			}
			for replicaIndex, _ := range module.Replicas {
				play, err := o.setupBareMetalJobPlay(newAppInstance, moduleIndex, replicaIndex, newModuleAction, commonInventoryStr, extraGlobalVars, *app)
				if err != nil {
					log.Error(err)
					return nil, err
				}

				plays = append(plays, play)
			}
		}
	case core.AppPlatformK8s:
		// 对于部署于k8s平台上的应用实例，所有的模块都使用[k8s-master]

		// 获取k8s集群主节点
		masterHost, err := o.helper.V1.K8sConfig.GetFirstMasterHost(oldAppInstance.Spec.K8sRef)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		// 生成旧版本卸载Play
		extraGlobalVars["package_dir"], _ = filepath.Abs(filepath.Join(setting.PackageSetting.PkgPath, oldVersionApp.PkgRef))
		for moduleIndex, module := range oldAppInstance.Spec.Modules {
			oldModuleAction, ok := oldModuleActions[module.Name]
			if !ok {
				continue
			}
			for replicaIndex, _ := range module.Replicas {
				play, err := o.setupK8sJobPlay(oldAppInstance, moduleIndex, replicaIndex, oldModuleAction, commonInventoryStr, extraGlobalVars, *app, masterHost)
				if err != nil {
					log.Error(err)
					return nil, err
				}

				plays = append(plays, play)
			}
		}

		// 生成新版本安装Play
		extraGlobalVars["package_dir"], _ = filepath.Abs(filepath.Join(setting.PackageSetting.PkgPath, newVersionApp.PkgRef))
		for moduleIndex, module := range newAppInstance.Spec.Modules {
			newModuleAction, ok := newModuleActions[module.Name]
			if !ok {
				continue
			}
			for replicaIndex, _ := range module.Replicas {
				play, err := o.setupK8sJobPlay(newAppInstance, moduleIndex, replicaIndex, newModuleAction, commonInventoryStr, extraGlobalVars, *app, masterHost)
				if err != nil {
					log.Error(err)
					return nil, err
				}

				plays = append(plays, play)
			}
		}
	}

	// 创建任务
	job := v2.NewJob()
	job.Metadata.Namespace = o.namespace
	job.Metadata.Name = fmt.Sprintf("%s-%s-%s-%s-to-%s-%d", core.KindAppInstance, newAppInstance.Metadata.Name, core.EventActionUpgrade, oldVersionApp.Version, newVersionApp.Version, time.Now().Unix())
	job.Spec.Exec.Type = core.JobExecTypeAnsible
	job.Spec.Exec.Ansible.Bin = setting.AnsibleSetting.Bin
	job.Spec.Exec.Ansible.Plays = plays
	job.Spec.TimeoutSeconds = 3600
	job.Spec.FailureThreshold = 1
	if _, err := o.helper.V2.Job.Create(context.TODO(), job); err != nil {
		log.Error(err)
		return nil, err
	}
	return job, nil
}

// 构建裸机任务play
// 一个play对应着应用实例中的一个模块副本的一个action，并且应用实例的每个模块可以对应着不同的应用版本
func (o AppInstanceOperator) setupBareMetalJobPlay(appInstance *v2.AppInstance, moduleIndex int, replicaIndex int, action string, commonInventoryStr string, extraGlobalVars map[string]interface{}, app v1.App) (v2.JobAnsiblePlay, error) {
	var play v2.JobAnsiblePlay

	module := appInstance.Spec.Modules[moduleIndex]
	replica := module.Replicas[replicaIndex]

	if module.AppVersion == "" {
		module.AppVersion = appInstance.Spec.AppRef.Version
	}

	versionApp, ok := app.GetVersion(module.AppVersion)
	if !ok {
		err := e.Errorf("version %s is not found in app %s", module.AppVersion, app.Metadata.Name)
		log.Error(err)
		return play, err
	} else if !versionApp.Enabled {
		err := e.Errorf("version %s is disabled in app %s", module.AppVersion, app.Metadata.Name)
		log.Error(err)
		return play, err
	}
	appModule, ok := versionApp.GetModule(module.Name)
	if !ok {
		err := e.Errorf("module %s is not found in app %s-%s", module.Name, app.Metadata.Name, module.AppVersion)
		log.Error(err)
		return play, err
	}

	inventory := make(map[string]ansible.InventoryGroup)
	playbook := ansible.Playbook{}
	configs := []v2.AnsibleConfig{}
	playName := fmt.Sprintf("%s-%d-%s", module.Name, replicaIndex, strings.ToLower(action))
	tags := []string{strings.ToLower(action)}

	groupVars := ansible.AppArgs{}

	// 生成模块参数
	for _, arg := range replica.Args {
		for _, appModule := range versionApp.Modules {
			if appModule.Name == module.Name {
				for _, appArg := range appModule.Args {
					if appArg.Name == arg.Name {
						// 根据参数类型填充参数值
						switch appArg.Type {
						case ArgTypeInteger:
							// 由于json反序列化会将数字类型统一转为float64, 因此需要先用类型推断转为float64再转为int
							var value interface{}
							switch v := arg.Value.(type) {
							case float64:
								switch appArg.Format {
								case ArgFormatInt32:
									value = int32(v)
								case ArgFormatInt64:
									value = int64(v)
								case ArgFormatPort:
									value = uint16(v)
								default:
									value = int64(v)
								}
							}
							groupVars.Set(arg.Name, value)
						case ArgTypeNumber:
							var value interface{}
							switch v := arg.Value.(type) {
							case float64:
								switch appArg.Format {
								case ArgFormatFloat:
									value = float32(v)
								case ArgFormatDouble:
									value = float64(v)
								default:
									value = float64(v)
								}
							}
							groupVars.Set(arg.Name, value)
						case ArgTypeBoolean:
							var value bool
							switch v := arg.Value.(type) {
							case bool:
								value = v
							}
							groupVars.Set(arg.Name, value)
						case ArgTypeString:
							var value interface{}
							switch v := arg.Value.(type) {
							case string:
								switch appArg.Format {
								case ArgFormatDate:
									value, _ = time.Parse(time.RFC3339, v)
									groupVars.Set(arg.Name, value)
								case ArgFormatArray:
									groupVars.Set(arg.Name, strings.Split(v, ";"))
								case ArgFormatGroupHost:
									hostRefs := strings.Split(v, ";")
									inventoryGroupHosts := make(map[string]ansible.InventoryHost)
									for _, hostRef := range hostRefs {
										hostObj, err := o.helper.V1.Host.Get(context.TODO(), o.namespace, hostRef)
										if err != nil {
											err = e.Errorf("failed to get host %s, %s", hostRef, err.Error())
											log.Error(err)
											return play, err
										} else if hostObj == nil {
											err := e.Errorf("host %s not found", hostRef)
											log.Error(err)
											return play, err
										}
										host := hostObj.(*v1.Host)
										inventoryGroupHosts[hostRef] = ansible.InventoryHost{
											"ansible_ssh_host": host.Spec.SSH.Host,
											"ansible_ssh_pass": host.Spec.SSH.Password,
											"ansible_ssh_port": host.Spec.SSH.Port,
											"ansible_ssh_user": host.Spec.SSH.User,
										}
									}
									inventory[arg.Name] = ansible.InventoryGroup{
										Hosts: inventoryGroupHosts,
										Vars:  make(map[string]interface{}),
									}
								default:
									groupVars.Set(arg.Name, v)
								}
							}
						default:
							groupVars.Set(arg.Name, arg.Value)
						}
						break
					}
				}
			}
		}
	}

	// 追加全局参数
	for _, arg := range appInstance.Spec.Global.Args {
		groupVars.Set(arg.Name, arg.Value)
	}

	// 追加额外组参数，都认定为内部参数
	for varName, varValue := range extraGlobalVars {
		groupVars[varName] = varValue
	}
	for varName, varValue := range appModule.ExtraVars {
		groupVars[varName] = varValue
	}
	playbook = ansible.Playbook{
		Hosts:       []string{module.Name},
		Roles:       appModule.IncludeRoles,
		IncludeVars: []string{"group_vars.yml"},
	}
	groupVars["module_name"] = module.Name
	groupVars["replica_index"] = replicaIndex
	groupVars["configs_dir"] = ansible.ConfigsDir
	groupVars["additional_configs_dir"] = ansible.AdditionalConfigsDir

	configs = append(configs, v2.AnsibleConfig{
		PathPrefix: ansible.ConfigsDir,
		ValueFrom: v2.ValueFrom{
			ConfigMapRef: replica.ConfigMapRef,
		},
	})

	configs = append(configs, v2.AnsibleConfig{
		PathPrefix: ansible.AdditionalConfigsDir,
		ValueFrom: v2.ValueFrom{
			ConfigMapRef: replica.AdditionalConfigMapRef,
		},
	})

	// 填充算法实例参数
	requestGPU := false
	supportGPUModels := []string{}
	if appModule.Resources.AlgorithmPlugin {
		var pluginName, pluginVersion, mediaType string
		for _, arg := range replica.Args {
			switch arg.Name {
			case "ALGORITHM_PLUGIN_NAME":
				switch v := arg.Value.(type) {
				case string:
					pluginName = v
				}
			case "ALGORITHM_PLUGIN_VERSION":
				switch v := arg.Value.(type) {
				case string:
					pluginVersion = v
				}
			case "ALGORITHM_MEDIA_TYPE":
				switch v := arg.Value.(type) {
				case string:
					mediaType = v
				}
			case "ALGORITHM_REQUEST_GPU":
				switch v := arg.Value.(type) {
				case bool:
					requestGPU = v
				}
			}
		}
		groupVars.Set("ALGORITHM_PLUGIN_NAME", pluginName)
		groupVars.Set("ALGORITHM_PLUGIN_VERSION", pluginVersion)
		groupVars.Set("ALGORITHM_MEDIA_TYPE", mediaType)

		// 填充算法插件参数
		pluginObj, err := o.helper.V1.App.Get(context.TODO(), core.DefaultNamespace, pluginName)
		if err != nil {
			log.Error(err)
			return play, e.Errorf("failed to get algorithm plugin %s", pluginName)
		}
		if pluginObj == nil {
			err := e.Errorf("algorithm plugin %s not found", pluginName)
			log.Error(err)
			return play, err
		}
		plugin := pluginObj.(*v1.App)
		for _, versionPlugin := range plugin.Spec.Versions {
			if versionPlugin.Version == pluginVersion {
				ap_package_dir, _ := filepath.Abs(filepath.Join(setting.PackageSetting.PkgPath, versionPlugin.PkgRef))
				groupVars["ap_package_dir"] = ap_package_dir

				for varKey, varValue := range versionPlugin.Modules[0].ExtraVars {
					groupVars[varKey] = varValue
				}
				supportGPUModels = versionPlugin.SupportGpuModels
			}
		}
	}

	// 构建当前模块副本的inventory
	inventoryGroupHosts := make(map[string]ansible.InventoryHost)
	algorithmGPUIDs := make(map[string]interface{})
	for _, hostRef := range replica.HostRefs {
		hostObj, err := o.helper.V1.Host.Get(context.TODO(), o.namespace, hostRef)
		if err != nil {
			err := e.Errorf("failed to get referred host %s: %s", hostRef, err.Error())
			log.Error(err)
			return play, err
		} else if hostObj == nil {
			err := e.Errorf("host %s not found", hostRef)
			log.Error(err)
			return play, err
		}
		host := hostObj.(*v1.Host)

		// 构建group hosts
		inventoryGroupHosts[hostRef] = ansible.InventoryHost{
			"ansible_ssh_host": host.Spec.SSH.Host,
			"ansible_ssh_pass": host.Spec.SSH.Password,
			"ansible_ssh_port": host.Spec.SSH.Port,
			"ansible_ssh_user": host.Spec.SSH.User,
			// 额外的主机参数
			"NODE_NAME": host.Metadata.Name,
		}

		// 在每台主机上寻找型号匹配且空闲的GPU与实例绑定
		if appModule.Resources.AlgorithmPlugin && requestGPU {
			gpuID, err := o.allocGPUSlot(appInstance, module.Name, replicaIndex, host, supportGPUModels, action)
			if err != nil {
				log.Error(err)
				return play, err
			}
			algorithmGPUIDs[hostRef] = gpuID
		}
	}
	groupVars["algorithm_gpu_ids"] = algorithmGPUIDs
	inventory[module.Name] = ansible.InventoryGroup{
		Hosts: inventoryGroupHosts,
	}
	// 添加额外的[k8s-master]，将裸机服务注册到k8s集群中
	if appInstance.Spec.K8sRef != "" {
		k8sObj, err := o.helper.V1.K8sConfig.Get(context.TODO(), o.namespace, appInstance.Spec.K8sRef)
		if err != nil {
			err = e.Errorf("failed to get k8s cluster %s, %s", appInstance.Spec.K8sRef, err.Error())
			log.Error(err)
			return play, err
		} else if k8sObj == nil {
			err := e.Errorf("k8s cluster '%s' not found", appInstance.Spec.K8sRef)
			log.Error(err)
			return play, err
		}
		k8s := k8sObj.(*v1.K8sConfig)

		k8sGroupHosts := make(map[string]ansible.InventoryHost)
		for _, k8sHost := range k8s.Spec.K8SMaster.Hosts {
			hostRef := k8sHost.ValueFrom.HostRef
			hostObj, err := o.helper.V1.Host.Get(context.TODO(), o.namespace, hostRef)
			if err != nil {
				err = e.Errorf("failed to get host %s, %s", hostRef, err.Error())
				log.Error(err)
				return play, err
			} else if hostObj == nil {
				err := e.Errorf("host %s not found", hostRef)
				log.Error(err)
				return play, err
			}
			host := hostObj.(*v1.Host)

			k8sGroupHosts[hostRef] = ansible.InventoryHost{
				"ansible_ssh_host": host.Spec.SSH.Host,
				"ansible_ssh_pass": host.Spec.SSH.Password,
				"ansible_ssh_user": host.Spec.SSH.User,
				"ansible_ssh_port": host.Spec.SSH.Port,
			}
		}

		inventory[ansible.ANSIBLE_GROUP_K8S_MASTER] = ansible.InventoryGroup{
			Hosts: k8sGroupHosts,
		}
	}

	// 序列化group_vars与inventory
	groupVarsData, _ := yaml.Marshal(groupVars)
	inventoryStr, _ := ansible.RenderInventory(inventory)
	playbookStr, _ := ansible.RenderPlaybook([]ansible.Playbook{playbook})

	play.Name = playName
	play.Configs = configs
	play.Envs = []string{"act=" + strings.ToLower(action)}
	play.Tags = tags
	play.GroupVars = v2.AnsibleGroupVars{
		Value: string(groupVarsData),
	}
	play.Inventory = v2.AnsibleInventory{
		Value: inventoryStr + commonInventoryStr,
	}
	play.Playbook = v2.AnsiblePlaybook{
		Value: playbookStr,
	}
	return play, nil
}

// 构建k8s任务play
// 一个play对应着应用实例中的一个模块副本的一个action，并且应用实例的每个模块可以对应着不同的应用版本
func (o AppInstanceOperator) setupK8sJobPlay(appInstance *v2.AppInstance, moduleIndex int, replicaIndex int, action string, commonInventoryStr string, extraGlobalVars map[string]interface{}, app v1.App, host *v1.Host) (v2.JobAnsiblePlay, error) {
	var play v2.JobAnsiblePlay

	module := appInstance.Spec.Modules[moduleIndex]
	replica := module.Replicas[replicaIndex]

	versionApp, ok := app.GetVersion(module.AppVersion)
	if !ok {
		err := e.Errorf("version %s is not found in app %s", module.AppVersion, app.Metadata.Name)
		log.Error(err)
		return play, err
	} else if !versionApp.Enabled {
		err := e.Errorf("version %s is disabled in app %s", module.AppVersion, app.Metadata.Name)
		log.Error(err)
		return play, err
	}
	appModule, ok := versionApp.GetModule(module.Name)
	if !ok {
		err := e.Errorf("module %s is not found in app %s-%s", module.Name, app.Metadata.Name, module.AppVersion)
		log.Error(err)
		return play, err
	}

	inventory := map[string]ansible.InventoryGroup{
		"ansible.ANSIBLE_GROUP_K8S_MASTER": ansible.InventoryGroup{
			Hosts: map[string]ansible.InventoryHost{
				host.Metadata.Name: ansible.InventoryHost{
					"ansible_ssh_host": host.Spec.SSH.Host,
					"ansible_ssh_pass": host.Spec.SSH.Password,
					"ansible_ssh_port": host.Spec.SSH.Port,
					"ansible_ssh_user": host.Spec.SSH.User,
				},
			},
		},
	}
	groupVars := ansible.AppArgs{}
	playbook := ansible.Playbook{}
	configs := []v2.AnsibleConfig{}
	playName := fmt.Sprintf("%s-%d", module.Name, replicaIndex)
	tags := []string{strings.ToLower(action)}

	for _, arg := range replica.Args {
		for _, appModule := range versionApp.Modules {
			if appModule.Name == module.Name {
				for _, appArg := range appModule.Args {
					if appArg.Name == arg.Name {
						// 根据参数类型填充参数值
						switch appArg.Type {
						case ArgTypeInteger:
							// 由于json反序列化会将数字类型统一转为float64, 因此需要先用类型推断转为float64再转为int
							var value interface{}
							switch v := arg.Value.(type) {
							case float64:
								switch appArg.Format {
								case ArgFormatInt32:
									value = int32(v)
								case ArgFormatInt64:
									value = int64(v)
								case ArgFormatPort:
									value = uint16(v)
								default:
									value = int64(v)
								}
							}
							groupVars.Set(arg.Name, value)
						case ArgTypeNumber:
							var value interface{}
							switch v := arg.Value.(type) {
							case float64:
								switch appArg.Format {
								case ArgFormatFloat:
									value = float32(v)
								case ArgFormatDouble:
									value = float64(v)
								default:
									value = float64(v)
								}
							}
							groupVars.Set(arg.Name, value)
						case ArgTypeBoolean:
							var value bool
							switch v := arg.Value.(type) {
							case bool:
								value = v
							}
							groupVars.Set(arg.Name, value)
						case ArgTypeString:
							var value interface{}
							switch v := arg.Value.(type) {
							case string:
								switch appArg.Format {
								case ArgFormatDate:
									value, _ = time.Parse(time.RFC3339, v)
									groupVars.Set(arg.Name, value)
								case ArgFormatArray:
									groupVars.Set(arg.Name, strings.Split(v, ";"))
								case ArgFormatGroupHost:
									hostRefs := strings.Split(v, ";")
									inventoryGroupHosts := make(map[string]ansible.InventoryHost)
									for _, hostRef := range hostRefs {
										hostObj, err := o.helper.V1.Host.Get(context.TODO(), o.namespace, hostRef)
										if err != nil {
											err = e.Errorf("failed to get host %s, %s", hostRef, err.Error())
											log.Error(err)
											return play, err
										} else if hostObj == nil {
											err := e.Errorf("host %s not found", hostRef)
											log.Error(err)
											return play, err
										}
										host := hostObj.(*v1.Host)
										inventoryGroupHosts[hostRef] = ansible.InventoryHost{
											"ansible_ssh_host": host.Spec.SSH.Host,
											"ansible_ssh_pass": host.Spec.SSH.Password,
											"ansible_ssh_port": host.Spec.SSH.Port,
											"ansible_ssh_user": host.Spec.SSH.User,
										}
									}
									inventory[arg.Name] = ansible.InventoryGroup{
										Hosts: inventoryGroupHosts,
										Vars:  make(map[string]interface{}),
									}
								}
							}
						default:
							groupVars.Set(arg.Name, arg.Value)
						}
						break
					}
				}
			}
		}
	}

	// 追加全局参数
	for _, arg := range appInstance.Spec.Global.Args {
		groupVars.Set(arg.Name, arg.Value)
	}

	// 追加额外组参数，都认定为内部参数
	for varName, varValue := range extraGlobalVars {
		groupVars[varName] = varValue
	}
	for varName, varValue := range appModule.ExtraVars {
		groupVars[varName] = varValue
	}
	playbook = ansible.Playbook{
		Hosts:       []string{ansible.ANSIBLE_GROUP_K8S_MASTER},
		Roles:       appModule.IncludeRoles,
		IncludeVars: []string{"group_vars.yml"},
	}

	configs = append(configs, v2.AnsibleConfig{
		PathPrefix: ansible.ConfigsDir,
		ValueFrom: v2.ValueFrom{
			ConfigMapRef: replica.ConfigMapRef,
		},
	})

	configs = append(configs, v2.AnsibleConfig{
		PathPrefix: ansible.AdditionalConfigsDir,
		ValueFrom: v2.ValueFrom{
			ConfigMapRef: replica.AdditionalConfigMapRef,
		},
	})

	groupVarsData, _ := yaml.Marshal(groupVars)
	inventoryStr, _ := ansible.RenderInventory(inventory)
	playbookStr, _ := ansible.RenderPlaybook([]ansible.Playbook{playbook})

	play.Name = playName
	play.Configs = configs
	play.Envs = []string{"act=" + strings.ToLower(action)}
	play.Tags = tags
	play.GroupVars = v2.AnsibleGroupVars{
		Value: string(groupVarsData),
	}
	play.Inventory = v2.AnsibleInventory{
		Value: inventoryStr + commonInventoryStr,
	}
	play.Playbook = v2.AnsiblePlaybook{
		Value: playbookStr,
	}
	return play, nil
}

// 将应用实例与主机上的GPU绑定并返回
func (o AppInstanceOperator) allocGPUSlot(appInstance *v2.AppInstance, moduleName string, replicaIndex int, host *v1.Host, supportGPUModels []string, action string) (int, error) {
	switch action {
	case core.EventActionInstall:
		// 绑定GPU
		for _, gpuInfo := range host.Spec.Info.GPUs {
			// 校验GPU型号
			modelMatched := false
			for _, model := range supportGPUModels {
				if model == gpuInfo.Type {
					modelMatched = true
					break
				}
			}
			if !modelMatched {
				continue
			}

			// 如果GPU未被使用，则绑定GPU
			gpuName := o.helper.V1.GPU.GetGPUName(host.Metadata.Name, gpuInfo.ID)
			gpuObj, err := o.helper.V1.GPU.Get(context.TODO(), "", gpuName)
			if err != nil {
				log.Error(err)
				return -1, err
			}
			if gpuObj == nil {
				err := e.Errorf("gpu %s not found", gpuName)
				log.Error(err)
				return -1, err
			}
			gpu := gpuObj.(*v1.GPU)
			if gpu.Status.Phase != core.PhaseBound {
				gpu.Spec.AppInstanceModuleRef = v1.AppInstanceModuleRef{
					AppInstanceRef: v1.AppInstanceRef{
						Namespace: appInstance.Metadata.Namespace,
						Name:      appInstance.Metadata.Name,
					},
					Module:  moduleName,
					Replica: replicaIndex,
				}
				gpu.Status.Phase = core.PhaseBound
				if _, err := o.helper.V1.GPU.Update(context.TODO(), gpu, core.WithStatus()); err != nil {
					log.Error(err)
					return -1, err
				}
				return gpuInfo.ID, nil
			}
		}
	case core.EventActionUninstall, core.EventActionConfigure:
		// 获取已绑定的GPU
		gpuObjs, err := o.helper.V1.GPU.List(context.TODO(), "")
		if err != nil {
			log.Error(err)
			return -1, err
		}
		for _, gpuObj := range gpuObjs {
			gpu := gpuObj.(*v1.GPU)
			if gpu.Spec.AppInstanceModuleRef.Namespace == appInstance.Metadata.Namespace && gpu.Spec.AppInstanceModuleRef.Name == appInstance.Metadata.Name && gpu.Spec.AppInstanceModuleRef.Module == moduleName && gpu.Spec.AppInstanceModuleRef.Replica == replicaIndex {
				return gpu.Spec.Info.ID, nil
			}
		}
	}
	return -1, e.Errorf("host %s is not bound with gpu types %v", host.Metadata.Name, supportGPUModels)
}

// 应用实例升级
func (o AppInstanceOperator) upgradeAppInstance(ctx context.Context, obj core.ApiObject) {
	newAppInstance := obj.(*v2.AppInstance)

	oldAppInstance := v2.NewAppInstance()
	if err := newAppInstance.DeepCopyInto(oldAppInstance); err != nil {
		log.Error(err)
		return
	}
	oldSpec := newAppInstance.Metadata.Annotations[core.AnnotationPrefix+"upgrade/last-applied-configuration"]
	err := oldAppInstance.SpecDecode([]byte(oldSpec))
	if err != nil {
		log.Error(err)
		return
	}
	job, err := o.setupUpgradeJob(oldAppInstance, newAppInstance)
	if err != nil {
		log.Error(err)
		return
	}

	eventMsg := "从 " + oldAppInstance.Spec.AppRef.Version + " 到 " + newAppInstance.Spec.AppRef.Version
	// 记录事件开始
	if err := o.recordEvent(Event{
		BaseApiObj: newAppInstance.BaseApiObj,
		Action:     core.EventActionUpgrade,
		Msg:        eventMsg,
		JobRef:     job.Metadata.Name,
		Phase:      core.PhaseWaiting,
	}); err != nil {
		log.Error(err)
	}
	// 监听升级job
	o.watchAndHandleJob(ctx, job.Metadata.Name, func(job *v2.Job) bool {
		switch job.Status.Phase {
		case core.PhaseWaiting, core.PhaseRunning:
			// 任务运行中，不做任何处理
			return false
		case core.PhaseCompleted:
			// 记录事件完成
			if err := o.recordEvent(Event{
				BaseApiObj: newAppInstance.BaseApiObj,
				Action:     core.EventActionUpgrade,
				Msg:        eventMsg,
				JobRef:     job.Metadata.Name,
				Phase:      core.PhaseCompleted,
			}); err != nil {
				log.Error(err)
			}
			delete(newAppInstance.Metadata.Annotations, core.AnnotationPrefix+"upgrade/last-applied-configuration")

			// 如果任务执行成功, 将应用实例置为Installed状态
			newAppInstance.Status.SetCondition(core.ConditionTypeInstalled, core.ConditionStatusTrue)
			newAppInstance.SetStatusPhase(core.PhaseInstalled)

			if err, _ := o.helper.V2.AppInstance.Update(context.TODO(), newAppInstance, core.WithStatus()); err != nil {
				log.Error(err)
			}
			return true
		case core.PhaseFailed:
			o.failback(newAppInstance, core.EventActionUpgrade, eventMsg, job)
			return true
		default:
			log.Warnf("unknown status phase '%s' of job '%s'", job.Status.Phase, job.GetKey())
			return false
		}
	})
}

// 应用实例升级回退
func (o AppInstanceOperator) upgradeBackoffAppInstance(ctx context.Context, obj core.ApiObject) {
	newAppInstance := obj.(*v2.AppInstance)

	oldAppInstance := v2.NewAppInstance()
	if err := newAppInstance.DeepCopyInto(oldAppInstance); err != nil {
		log.Error(err)
		return
	}
	oldSpec := newAppInstance.Metadata.Annotations[core.AnnotationPrefix+"upgrade/last-applied-configuration"]
	err := oldAppInstance.SpecDecode([]byte(oldSpec))
	if err != nil {
		log.Error(err)
		return
	}
	job, err := o.setupUpgradeJob(newAppInstance, oldAppInstance)
	if err != nil {
		log.Error(err)
		return
	}

	eventMsg := "从 " + newAppInstance.Spec.AppRef.Version + " 到 " + oldAppInstance.Spec.AppRef.Version
	// 记录事件开始
	if err := o.recordEvent(Event{
		BaseApiObj: newAppInstance.BaseApiObj,
		Action:     core.EventActionUpgradeBackoff,
		Msg:        eventMsg,
		JobRef:     job.Metadata.Name,
		Phase:      core.PhaseWaiting,
	}); err != nil {
		log.Error(err)
	}
	// 监听升级job
	o.watchAndHandleJob(ctx, job.Metadata.Name, func(job *v2.Job) bool {
		switch job.Status.Phase {
		case core.PhaseWaiting, core.PhaseRunning:
			// 任务运行中，不做任何处理
			return false
		case core.PhaseCompleted:
			// 记录事件完成
			if err := o.recordEvent(Event{
				BaseApiObj: newAppInstance.BaseApiObj,
				Action:     core.EventActionUpgrade,
				Msg:        eventMsg,
				JobRef:     job.Metadata.Name,
				Phase:      core.PhaseCompleted,
			}); err != nil {
				log.Error(err)
			}
			delete(newAppInstance.Metadata.Annotations, core.AnnotationPrefix+"upgrade/last-applied-configuration")

			// 如果任务执行成功, 将应用实例置为Installed状态
			oldAppInstance.Status.SetCondition(core.ConditionTypeInstalled, core.ConditionStatusTrue)
			oldAppInstance.SetStatusPhase(core.PhaseInstalled)

			if err, _ := o.helper.V2.AppInstance.Update(context.TODO(), oldAppInstance, core.WithStatus()); err != nil {
				log.Error(err)
			}
			return true
		case core.PhaseFailed:
			o.failback(newAppInstance, core.EventActionUpgradeBackoff, eventMsg, job)
			return true
		default:
			log.Warnf("unknown status phase '%s' of job '%s'", job.Status.Phase, job.GetKey())
			return false
		}
	})
}

func (o AppInstanceOperator) releaseGPU(appInstance *v2.AppInstance) error {
	gpuObjs, err := o.helper.V1.GPU.List(context.TODO(), "")
	if err != nil {
		log.Error(err)
		return err
	}
	for _, gpuObj := range gpuObjs {
		gpu := gpuObj.(*v1.GPU)
		if gpu.Spec.AppInstanceModuleRef.Namespace == appInstance.Metadata.Namespace && gpu.Spec.AppInstanceModuleRef.Name == appInstance.Metadata.Name {
			gpu.Spec.AppInstanceModuleRef = v1.AppInstanceModuleRef{}
			gpu.Status.Phase = core.PhaseWaiting
			log.Debugf("%+v", gpu)
			if _, err := o.helper.V1.GPU.Update(context.TODO(), gpu, core.WithStatus()); err != nil {
				log.Error(err)
				return err
			}
		}
	}
	return nil
}

func NewAppInstanceOperator() *AppInstanceOperator {
	o := &AppInstanceOperator{
		BaseOperator: NewBaseOperator(v2.NewAppInstanceRegistry()),
	}
	o.SetHandleFunc(o.handleAppInstance)
	o.healthCheckMap = make(map[string]context.CancelFunc)
	return o
}
