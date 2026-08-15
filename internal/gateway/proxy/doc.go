// Package proxy 实现 SSE 流式透明代理，按 chunk 透传上游响应并在单连接缓冲超限时施加 TCP 背压。
package proxy
