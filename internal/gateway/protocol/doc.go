// Package protocol 协议识别：TLS SNI、HTTP/1.1、HTTP/2、SSE 分帧。
// 识别不出来的协议默认透传 + 打告警（fail-closed 可单独配）。
package protocol
