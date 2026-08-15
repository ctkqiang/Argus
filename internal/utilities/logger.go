package utilities

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// LogLevel 定义日志级别枚举。
// 字符串形式便于通过配置文件 / 环境变量直接解析，
// 严格递增顺序为 Debug < Info < Warn < Error < Fatal。
type LogLevel string

const (
	// Debug 用于开发调试细节，生产默认关闭。启用后会自动附带源码文件与行号。
	Debug LogLevel = "debug"
	// Info 用于正常业务流程的里程碑事件。高吞吐路径下不应滥用，避免日志洪水。
	Info LogLevel = "info"
	// Warn 用于可恢复的非预期情况，请求本身仍能正常完成。
	// 例如 "第 2 次重试"、"证书将在 3 天后过期"。
	Warn LogLevel = "warn"
	// Error 用于单次请求或单条操作失败，但应用整体仍然健康。
	Error LogLevel = "error"
	// Fatal 用于应用无法继续运行的致命错误（仅限启动阶段），
	// 输出后立即以退出码 1 终止进程。业务请求路径禁止使用。
	Fatal LogLevel = "fatal"
)

// Format 描述日志输出的编码格式。
type Format string

const (
	// FmtText 采用 `key=value` 的人类可读文本形式，适合本地开发与 `tail -f` 场景。
	FmtText Format = "text"
	// FmtJSON 采用每行一条 JSON 的结构化格式，适合 ELK / Loki / Datadog 等日志平台直接消费。
	FmtJSON Format = "json"
)

// fatalSlogLevel 将自定义的 Fatal 级别映射到 slog 内置 Level 之上的自定义数值。
const fatalSlogLevel = slog.LevelError + 4

// levelToSlog 将字符串形式的 LogLevel 转换为 slog.Level。
// 未识别的值默认降级为 Info，以避免配置写错时日志消失。
func levelToSlog(l LogLevel) slog.Level {
	switch l {
	case Debug:
		return slog.LevelDebug
	case Info:
		return slog.LevelInfo
	case Warn:
		return slog.LevelWarn
	case Error:
		return slog.LevelError
	case Fatal:
		return fatalSlogLevel
	default:
		return slog.LevelInfo
	}
}

// replaceFatalLevel 是 slog Handler 的 ReplaceAttr 钩子，
// 负责把自定义数值级别的 fatalSlogLevel 渲染成字符串 "fatal"。
func replaceFatalLevel(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		level, ok := a.Value.Any().(slog.Level)
		if ok && level == fatalSlogLevel {
			a.Value = slog.StringValue(string(Fatal))
		}
	}
	return a
}

// Logger 是 Argus 项目统一的结构化日志入口。
// 底层使用标准库 log/slog 实现，零外部依赖。
// 同一 Logger 上所有方法均并发安全；With 方法返回的派生 Logger
// 与原始 Logger 共享级别控制器，SetLevel 对所有派生实例同步生效。
type Logger struct {
	inner    *slog.Logger
	levelVar *slog.LevelVar
	out      io.Writer
	level    LogLevel
	fmt      Format
}

// Option 是构造 Logger 的函数式选项，保持构造 API 稳定且向后兼容。
// 新配置项一律通过新增 WithXxx 函数引入，不修改 NewLogger 签名。
type Option func(*Logger)

// WithLevel 设置最低输出级别，低于该级别的日志会被静默丢弃。
func WithLevel(level LogLevel) Option {
	return func(l *Logger) { l.level = level }
}

// WithOutput 设置日志写入目标，典型取值：os.Stderr、os.Stdout 或任意 io.Writer。
func WithOutput(w io.Writer) Option {
	return func(l *Logger) { l.out = w }
}

// WithFormat 设置日志输出格式（FmtText 或 FmtJSON）。
func WithFormat(f Format) Option {
	return func(l *Logger) { l.fmt = f }
}

// NewLogger 根据可变 Option 构造新的 Logger 实例。
// 默认配置：级别 Info、输出 os.Stderr、格式 FmtText。
// 当级别设为 Debug 时，会自动开启源码定位 (AddSource)。
func NewLogger(opts ...Option) *Logger {
	l := &Logger{
		level:    Info,
		out:      os.Stderr,
		fmt:      FmtText,
		levelVar: new(slog.LevelVar),
	}
	for _, opt := range opts {
		opt(l)
	}
	l.buildInner()
	return l
}

// buildInner 根据当前配置重建底层 slog.Handler。
// 仅在 format / output 切换时调用；级别变更走 levelVar 动态更新。
func (l *Logger) buildInner() {
	l.levelVar.Set(levelToSlog(l.level))

	handlerOpts := &slog.HandlerOptions{
		Level:       l.levelVar,
		AddSource:   l.level == Debug,
		ReplaceAttr: replaceFatalLevel,
	}

	var handler slog.Handler
	if l.fmt == FmtJSON {
		handler = slog.NewJSONHandler(l.out, handlerOpts)
	} else {
		handler = slog.NewTextHandler(l.out, handlerOpts)
	}
	l.inner = slog.New(handler)
}

// Level 返回当前最低输出级别。
func (l *Logger) Level() LogLevel { return l.level }

// Format 返回当前日志输出格式。
func (l *Logger) Format() Format { return l.fmt }

// SetLevel 动态调整最低输出级别。所有通过 With 派生的子 Logger 会同步生效，
// 因为共享同一个 levelVar。该方法并发安全。
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
	l.levelVar.Set(levelToSlog(level))
}

// SetFormat 切换日志输出格式。需要重建 Handler，所以调用成本略高于 SetLevel。
func (l *Logger) SetFormat(f Format) {
	l.fmt = f
	l.buildInner()
}

// SetOutput 切换日志输出目标。需要重建 Handler。
func (l *Logger) SetOutput(w io.Writer) {
	l.out = w
	l.buildInner()
}

// LogContext 是核心输出方法：携带 context、级别、消息与可变 KV 字段。
// 所有 Log* 与 Log*Context 快捷方法均薄包装到此方法，保持唯一真实逻辑点。
// Fatal 级别会在输出后调用 os.Exit(1)。
func (l *Logger) LogContext(ctx context.Context, level LogLevel, msg string, args ...any) {
	l.inner.Log(ctx, levelToSlog(level), msg, args...)
	if level == Fatal {
		os.Exit(1)
	}
}

// Log 以 context.Background() 调用 LogContext。适合快捷调用或不关心取消语义的场景。
func (l *Logger) Log(level LogLevel, msg string, args ...any) {
	l.LogContext(context.Background(), level, msg, args...)
}

// LogDebugContext 输出 Debug 级别日志，携带调用方 context。
func (l *Logger) LogDebugContext(ctx context.Context, msg string, args ...any) {
	l.LogContext(ctx, Debug, msg, args...)
}

// LogDebug 等价于 LogDebugContext(context.Background(), ...)。
func (l *Logger) LogDebug(msg string, args ...any) {
	l.LogContext(context.Background(), Debug, msg, args...)
}

// LogInfoContext 输出 Info 级别日志，携带调用方 context。
func (l *Logger) LogInfoContext(ctx context.Context, msg string, args ...any) {
	l.LogContext(ctx, Info, msg, args...)
}

// LogInfo 等价于 LogInfoContext(context.Background(), ...)。
func (l *Logger) LogInfo(msg string, args ...any) {
	l.LogContext(context.Background(), Info, msg, args...)
}

// LogWarnContext 输出 Warn 级别日志，携带调用方 context。
func (l *Logger) LogWarnContext(ctx context.Context, msg string, args ...any) {
	l.LogContext(ctx, Warn, msg, args...)
}

// LogWarn 等价于 LogWarnContext(context.Background(), ...)。
func (l *Logger) LogWarn(msg string, args ...any) {
	l.LogContext(context.Background(), Warn, msg, args...)
}

// LogErrorContext 输出 Error 级别日志，携带调用方 context。
func (l *Logger) LogErrorContext(ctx context.Context, msg string, args ...any) {
	l.LogContext(ctx, Error, msg, args...)
}

// LogError 等价于 LogErrorContext(context.Background(), ...)。
func (l *Logger) LogError(msg string, args ...any) {
	l.LogContext(context.Background(), Error, msg, args...)
}

// LogFatalContext 输出 Fatal 级别日志并以退出码 1 终止进程。
// 仅限 main 启动阶段使用；请求路径严禁调用。
func (l *Logger) LogFatalContext(ctx context.Context, msg string, args ...any) {
	l.LogContext(ctx, Fatal, msg, args...)
}

// LogFatal 等价于 LogFatalContext(context.Background(), ...)。
func (l *Logger) LogFatal(msg string, args ...any) {
	l.LogContext(context.Background(), Fatal, msg, args...)
}

// With 为 Logger 附加上下文固定字段（如 trace_id、user_id），返回新的 Logger 实例。
// 新旧实例共享 levelVar，因而动态调级始终同步。
// 该方法不可变（immutable）：不会修改接收者本身。
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		inner:    l.inner.With(args...),
		levelVar: l.levelVar,
		out:      l.out,
		level:    l.level,
		fmt:      l.fmt,
	}
}

// defaultLogger 是包级共享的默认实例。init 中轻量初始化，
// 并通过 Default / SetLevel / SetOutput / SetFormat 提供受控的可替换能力，
// 以便单元测试切换到 bytes.Buffer 等内存写入器。
var defaultLogger *Logger

func init() {
	defaultLogger = NewLogger()
}

// Default 返回包级默认 Logger。请优先依赖注入具体 Logger，仅在装配层或工具函数中使用本便捷实例。
func Default() *Logger { return defaultLogger }

// SetLevel 调整默认 Logger 的最低输出级别。对通过 With 派生的所有实例同步生效。
func SetLevel(level LogLevel) { defaultLogger.SetLevel(level) }

// SetOutput 切换默认 Logger 的输出目标。
func SetOutput(w io.Writer) { defaultLogger.SetOutput(w) }

// SetFormat 切换默认 Logger 的输出格式。
func SetFormat(f Format) { defaultLogger.SetFormat(f) }

// With 在默认 Logger 上附加固定字段，返回派生 Logger。
func With(args ...any) *Logger { return defaultLogger.With(args...) }

// Log 在默认 Logger 上输出指定级别的日志。
func Log(level LogLevel, msg string, args ...any) {
	defaultLogger.Log(level, msg, args...)
}

// LogContext 在默认 Logger 上输出指定级别的日志，携带调用方 context。
func LogContext(ctx context.Context, level LogLevel, msg string, args ...any) {
	defaultLogger.LogContext(ctx, level, msg, args...)
}
