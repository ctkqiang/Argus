// Package logger 是 Argus 对外日志入口。
// 仅作为 internal/utilities/logger 的薄封装导出，
// 保证第三方调用方（如果未来要引用 pkg）仍能拿到一致 API。
// 内部禁止引入三方日志库，底层实现始终基于 Go 标准库 log/slog。
//
// 注意：本包所有导出符号均为 internal/utilities/logger 的直接转发。
// 请优先依赖注入具体 Logger 实例，仅在装配层或工具函数中使用包级便捷函数。
package logger

import "github.com/ctkqiang/argus/internal/utilities"

// LogLevel 是 utilities.LogLevel 的类型别名，保留相同枚举语义。
// 字符串形式便于通过配置文件 / 环境变量直接解析，
// 严格递增顺序为 Debug < Info < Warn < Error < Fatal。
type LogLevel = utilities.LogLevel

// 日志级别枚举值。此处以变量形式转发 utilities 中的同名常量，
// 因为 Go 不支持对 const 做 type alias 以外的再导出。
// 调用方可直接使用 logger.Debug / logger.Info 等，与 utilities 包完全等价。
var (
	// Debug 用于开发调试细节，生产默认关闭。启用后会自动附带源码文件与行号。
	Debug = utilities.Debug
	// Info 用于正常业务流程的里程碑事件。高吞吐路径下不应滥用，避免日志洪水。
	Info = utilities.Info
	// Warn 用于可恢复的非预期情况，请求本身仍能正常完成。
	// 例如 "第 2 次重试"、"证书将在 3 天后过期"。
	Warn = utilities.Warn
	// Error 用于单次请求或单条操作失败，但应用整体仍然健康。
	Error = utilities.Error
	// Fatal 用于应用无法继续运行的致命错误（仅限启动阶段），
	// 输出后立即以退出码 1 终止进程。业务请求路径禁止使用。
	Fatal = utilities.Fatal
)

// Format 是 utilities.Format 的类型别名，描述日志输出的编码格式。
type Format = utilities.Format

// 日志格式枚举值。与级别同理，以变量形式转发。
var (
	// FmtText 采用 `key=value` 的人类可读文本形式，适合本地开发与 `tail -f` 场景。
	FmtText = utilities.FmtText
	// FmtJSON 采用每行一条 JSON 的结构化格式，适合 ELK / Loki / Datadog 等日志平台直接消费。
	FmtJSON = utilities.FmtJSON
)

// Logger 是 utilities.Logger 的类型别名。
// 同一 Logger 上所有方法均并发安全；With 方法返回的派生 Logger
// 与原始 Logger 共享级别控制器，SetLevel 对所有派生实例同步生效。
type Logger = utilities.Logger

// Option 是构造 Logger 的函数式选项类型别名。
// 新配置项一律通过 utilities 包新增 WithXxx 函数引入，不修改 NewLogger 签名。
type Option = utilities.Option

var (
	// WithLevel 设置最低输出级别，低于该级别的日志会被静默丢弃。
	WithLevel = utilities.WithLevel

	// WithOutput 设置日志写入目标，典型取值：os.Stderr、os.Stdout 或任意 io.Writer。
	WithOutput = utilities.WithOutput

	// WithFormat 设置日志输出格式（FmtText 或 FmtJSON）。
	WithFormat = utilities.WithFormat

	// NewLogger 根据可变 Option 构造新的 Logger 实例。
	// 默认配置：级别 Info、输出 os.Stderr、格式 FmtText。
	// 当级别设为 Debug 时，会自动开启源码定位 (AddSource)。
	NewLogger = utilities.NewLogger

	// Default 返回包级默认 Logger。请优先依赖注入具体 Logger，
	// 仅在装配层或工具函数中使用本便捷实例。
	Default = utilities.Default

	// SetLevel 调整默认 Logger 的最低输出级别。对通过 With 派生的所有实例同步生效。
	SetLevel = utilities.SetLevel

	// SetOutput 切换默认 Logger 的输出目标。
	SetOutput = utilities.SetOutput

	// SetFormat 切换默认 Logger 的输出格式。
	SetFormat = utilities.SetFormat

	// With 在默认 Logger 上附加固定字段（如 trace_id、user_id），返回派生 Logger。
	With = utilities.With

	// Log 在默认 Logger 上输出指定级别的日志。
	Log = utilities.Log

	// LogContext 在默认 Logger 上输出指定级别的日志，携带调用方 context。
	LogContext = utilities.LogContext
)
