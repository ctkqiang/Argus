// Package event gateway 侧的事件上报客户端。
// AIEvent 攒批 + 流式上报 + 失败重试 + 本地缓冲兜底，不阻塞请求热路径。
package event
