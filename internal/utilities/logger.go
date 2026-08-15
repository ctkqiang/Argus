package utilities

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// LogLevel 日志级别字符串。直接配成 env 或 YAML 都方便。
// 顺序：Debug < Info < Warn < Error < Fatal。
type LogLevel string

const (
	// Debug 开发细节。生产默认关，开了会自动带文件和行号。
	Debug LogLevel = "debug"
	// Info 正常流程里的关键节点。别乱打到热路径上把日志撑爆。
	Info LogLevel = "info"
	// Warn 可恢复的异常：比如第二次重试、证书还有三天过期。
	Warn LogLevel = "warn"
	// Error 某条操作失败，但整体程序还活着。
	Error LogLevel = "error"
	// Fatal 启动期才允许用。输出之后直接 os.Exit(1)。请求路径别碰。
	Fatal LogLevel = "fatal"
)

// Format 日志输出格式。
type Format string

const (
	// FmtText key=value 的人读文本。本地 tail -f 最舒服。
	FmtText Format = "text"
	// FmtJSON 每行一条 JSON。直接丢给 Loki / Datadog 之类的平台。
	FmtJSON Format = "json"
)

// fatalSlogLevel slog 没有 Fatal，我们自己加一段偏移。
const fatalSlogLevel = slog.LevelError + 4

// levelToSlog 把字符串级别映射到 slog.Level。写错了默认回 Info，别把日志直接弄没。
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

// replaceFatalLevel slog 的 ReplaceAttr 钩子，把我们自己的 fatal 级别渲染成 "fatal"。
func replaceFatalLevel(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		level, ok := a.Value.Any().(slog.Level)
		if ok && level == fatalSlogLevel {
			a.Value = slog.StringValue(string(Fatal))
		}
	}
	return a
}

// Logger 结构化日志入口，内部就是 slog。零外部依赖。
// With 出来的派生 logger 和原 logger 共享 levelVar，动态调级会一起变。
type Logger struct {
	inner    *slog.Logger
	levelVar *slog.LevelVar
	out      io.Writer
	level    LogLevel
	fmt      Format
}

// Option 给 NewLogger 传可选参数。以后要加配置就写新的 WithXxx，别改 NewLogger 签名。
type Option func(*Logger)

// WithLevel 设最低输出级别。低于这个级别的会被丢掉。
func WithLevel(level LogLevel) Option {
	return func(l *Logger) { l.level = level }
}

// WithOutput 设输出目标。os.Stderr / os.Stdout / 任意 io.Writer 都行。
func WithOutput(w io.Writer) Option {
	return func(l *Logger) { l.out = w }
}

// WithFormat 设输出格式：FmtText 或 FmtJSON。
func WithFormat(f Format) Option {
	return func(l *Logger) { l.fmt = f }
}

// NewLogger 构造 Logger。
// 默认：Info 级别、写 os.Stderr、text 格式；Debug 级自动加源码定位。
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

// buildInner 按当前 output/format 重建 slog.Handler。
// 级别切换不走这里，直接走 levelVar.Set。
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

// Level 当前最低级别。
func (l *Logger) Level() LogLevel { return l.level }

// Format 当前输出格式。
func (l *Logger) Format() Format { return l.fmt }

// SetLevel 动态切级别。levelVar 是 slog 自带并发安全的。
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
	l.levelVar.Set(levelToSlog(level))
}

// SetFormat 切格式。需要重建 Handler，比 SetLevel 稍重。
func (l *Logger) SetFormat(f Format) {
	l.fmt = f
	l.buildInner()
}

// SetOutput 切输出目标。需要重建 Handler。
func (l *Logger) SetOutput(w io.Writer) {
	l.out = w
	l.buildInner()
}

// LogContext 所有输出最终都走这里。Fatal 打完会 os.Exit(1)。
func (l *Logger) LogContext(ctx context.Context, level LogLevel, msg string, args ...any) {
	l.inner.Log(ctx, levelToSlog(level), msg, args...)
	if level == Fatal {
		os.Exit(1)
	}
}

// Log 没 context 的快捷版本，内部用 context.Background()。
func (l *Logger) Log(level LogLevel, msg string, args ...any) {
	l.LogContext(context.Background(), level, msg, args...)
}

// LogDebugContext Debug。带 ctx。
func (l *Logger) LogDebugContext(ctx context.Context, msg string, args ...any) {
	l.LogContext(ctx, Debug, msg, args...)
}

// LogDebug Debug。
func (l *Logger) LogDebug(msg string, args ...any) {
	l.LogContext(context.Background(), Debug, msg, args...)
}

// LogInfoContext Info。带 ctx。
func (l *Logger) LogInfoContext(ctx context.Context, msg string, args ...any) {
	l.LogContext(ctx, Info, msg, args...)
}

// LogInfo Info。
func (l *Logger) LogInfo(msg string, args ...any) {
	l.LogContext(context.Background(), Info, msg, args...)
}

// LogWarnContext Warn。带 ctx。
func (l *Logger) LogWarnContext(ctx context.Context, msg string, args ...any) {
	l.LogContext(ctx, Warn, msg, args...)
}

// LogWarn Warn。
func (l *Logger) LogWarn(msg string, args ...any) {
	l.LogContext(context.Background(), Warn, msg, args...)
}

// LogErrorContext Error。带 ctx。
func (l *Logger) LogErrorContext(ctx context.Context, msg string, args ...any) {
	l.LogContext(ctx, Error, msg, args...)
}

// LogError Error。
func (l *Logger) LogError(msg string, args ...any) {
	l.LogContext(context.Background(), Error, msg, args...)
}

// LogFatalContext Fatal。打完就 exit，只给启动装配用。
func (l *Logger) LogFatalContext(ctx context.Context, msg string, args ...any) {
	l.LogContext(ctx, Fatal, msg, args...)
}

// LogFatal Fatal。
func (l *Logger) LogFatal(msg string, args ...any) {
	l.LogContext(context.Background(), Fatal, msg, args...)
}

// With 给 logger 挂一组固定字段（比如 trace_id / user_id），返回新 logger。
// 接收者本身不变。共享 levelVar，所以动态调级同步生效。
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		inner:    l.inner.With(args...),
		levelVar: l.levelVar,
		out:      l.out,
		level:    l.level,
		fmt:      l.fmt,
	}
}

// defaultLogger 包级默认 logger。init 里轻量 new 一个。
// 之所以提供 SetOutput / SetFormat / SetLevel，是为了单元测试里
// 能切到 bytes.Buffer，不会把终端刷满。
var defaultLogger *Logger

func init() {
	defaultLogger = NewLogger()
}

// Default 默认 logger。能 DI 就 DI，图省事再用这个。
func Default() *Logger { return defaultLogger }

// SetLevel 调默认 logger 级别。
func SetLevel(level LogLevel) { defaultLogger.SetLevel(level) }

// SetOutput 切默认 logger 输出。
func SetOutput(w io.Writer) { defaultLogger.SetOutput(w) }

// SetFormat 切默认 logger 格式。
func SetFormat(f Format) { defaultLogger.SetFormat(f) }

// With 默认 logger 挂固定字段。
func With(args ...any) *Logger { return defaultLogger.With(args...) }

// Log 默认 logger 输出。
func Log(level LogLevel, msg string, args ...any) {
	defaultLogger.Log(level, msg, args...)
}

// LogContext 默认 logger 带 ctx 输出。
func LogContext(ctx context.Context, level LogLevel, msg string, args ...any) {
	defaultLogger.LogContext(ctx, level, msg, args...)
}
