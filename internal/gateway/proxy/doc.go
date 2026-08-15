// Package proxy 实现决策后的 SSE 流式反向代理，支持响应体脱敏掩码、
// token 级审计、流式错误注入。依赖 adapter 层进行协议转换，禁止直接调用上游 HTTP。
package proxy
