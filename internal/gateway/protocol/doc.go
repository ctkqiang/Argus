// Package protocol 实现 LLM API 协议识别与解析，支持 OpenAI、Anthropic、
// Azure OpenAI 等主流厂商的请求/响应格式归一化。解析结果作为 pipeline 的输入，
// 禁止对原始 payload 做任何修改。
package protocol
