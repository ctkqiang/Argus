// Package server 实现网关入口 HTTP/HTTPS 服务器，负责透明引流流量的监听、
// 连接管理、优雅启停与基础限流。仅依赖 Go 标准库 net/http 与 internal/utilities，
// 业务逻辑委托给 protocol 与 pipeline 层。
package server
