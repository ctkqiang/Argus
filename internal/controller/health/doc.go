// Package health 提供控制平面自身的健康探针与存活/就绪状态管理，
// 包含 leader-election 状态、CRD 缓存同步状态、下游数据库连接状态等。
// 通过 /healthz /readyz 端点暴露给 kubelet。
package health
