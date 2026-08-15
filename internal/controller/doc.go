// Package controller 实现观枢控制平面的核心逻辑，包含 CRD Watch 调谐、
// 策略下发、Pod 身份服务、AIEvent 持久化等控制器模块。该包基于 controller-runtime
// 框架构建，依赖 api/argus/v1alpha1 定义的 CRD 类型。
package controller
