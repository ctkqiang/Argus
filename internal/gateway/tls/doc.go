// Package tls TLS 终止：按 ClientHello SNI 选叶子证书。
// 私钥只从 Secret 加载，禁落盘、禁打日志、永远别 InsecureSkipVerify=true。
package tls
