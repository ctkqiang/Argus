// Package tls 提供网关侧 TLS 终止与 mTLS 双向认证能力，包含证书加载、
// SNI 路由、客户端证书校验等功能。密钥材料仅通过环境变量或 k8s Secret 注入，
// 禁止硬编码或写入日志。
package tls
