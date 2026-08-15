// Package tracing 提供 OpenTelemetry 分布式追踪初始化工具，
// 支持 OTLP gRPC/HTTP Exporter、采样率配置、Trace ID 注入到日志上下文。
// 该包可独立被外部项目引用，不依赖 Argus 内部模块。
package tracing
