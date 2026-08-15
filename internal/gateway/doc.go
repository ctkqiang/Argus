// Package gateway 数据平面的业务入口。
// 禁止反向依赖 internal/controller。要查身份、要拿策略，通通走 gRPC 问 controller。
package gateway
