// Package policy 实现 ArgusSecurityPolicy CRD 的调谐控制器，负责 Watch 策略变更、
// 校验规则合法性、编译为网关可消费的二进制格式并通过 ConfigMap/CRD 下发到数据平面。
// 支持策略版本管理与灰度发布。
package policy
