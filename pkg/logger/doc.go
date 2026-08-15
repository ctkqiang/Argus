// Package logger 是 internal/utilities/logger 的对外薄封装，
// 暴露与内部一致的结构化日志接口供下游集成方使用。实现上仅做类型转发，
// 保证内部日志行为与外部 SDK 行为严格一致。
package logger
