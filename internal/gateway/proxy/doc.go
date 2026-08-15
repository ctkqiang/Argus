// Package proxy 决策做完后的反向代理。
// SSE 按 chunk 转发，不聚合整包；慢客户端直接 TCP 反压，别把网关吃 OOM。
package proxy
