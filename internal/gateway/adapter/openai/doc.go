// Package openai 实现 OpenAI 兼容协议（含 Azure OpenAI、vLLM、Ollama OpenAI 模式）
// 的请求/响应适配器，支持 Chat Completions、Completions、Embeddings 三类端点。
// 仅依赖父包 adapter 定义的接口与 internal/utilities。
package openai
