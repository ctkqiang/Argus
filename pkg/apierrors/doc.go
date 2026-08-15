// Package apierrors 定义 Argus 对外 HTTP API 的标准化错误类型与错误码，
// 包含网关阻断响应、鉴权失败、限流、上游超时等场景。错误码命名遵循 HTTP 语义 + 业务子码，
// 错误体结构固定，便于上游 SDK 统一解析。
package apierrors
