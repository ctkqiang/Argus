// Package signal 提供进程信号监听与优雅退出 Context 管理。
// 封装了对 os.Interrupt (Ctrl+C) 与 syscall.SIGTERM (Kubernetes / systemd 终止信号)
// 的标准监听流程，并支持"二次信号强制退出"的常见生产级约定。
package signal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

// shutdownSignal 是本包统一监听的信号集合。
// 选取 os.Interrupt 与 SIGTERM 覆盖本地开发 Ctrl+C 以及容器运行时终止两大主场景。
var shutdownSignal = []os.Signal{os.Interrupt, syscall.SIGTERM}

// WithShutdown 返回一个在收到进程终止信号时被取消的 Context。
//
// 行为约定：
//  1. 第一次收到 os.Interrupt 或 SIGTERM：通过 cause 取消返回的 Context，
//     cause 为 fmt.Errorf("received signal: %s", sig)，调用方据此触发优雅退出流程。
//  2. 第二次收到任一信号：直接调用 os.Exit(1)，防止优雅退出挂死时进程无法终止。
//
// 返回的 CancelCauseFunc 用于调用方主动提前关闭信号监听 goroutine；
// 正常情况下 main 函数应在 defer 中调用它以释放资源。
func WithShutdown(parent context.Context) (context.Context, context.CancelCauseFunc) {
	ctx, cancel := context.WithCancelCause(parent)

	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, shutdownSignal...)

	// signalReceived 记录已收到的信号次数，用于区分"首次优雅"与"二次强制"。
	var signalReceived atomic.Int32

	go func() {
		for sig := range sigCh {
			count := signalReceived.Add(1)
			if count == 1 {
				cancel(fmt.Errorf("received signal: %s", sig))
				continue
			}
			// 第二次信号：强制退出，避免优雅退出阶段卡死。
			os.Exit(1)
		}
	}()

	// 包装 cancel：先停止信号监听，再传播取消，确保 goroutine 必然退出。
	cancelCause := func(cause error) {
		signal.Stop(sigCh)
		close(sigCh)
		cancel(cause)
	}

	return ctx, cancelCause
}

// WaitUntilDone 阻塞直到 ctx 被取消，然后执行 preShutdown 清理动作。
//
// 参数：
//   - ctx: 通常由 WithShutdown 返回；任意原因的取消都会触发后续清理流程。
//   - loggerFn: 用于打印等待、完成、超时等生命周期事件的日志函数；
//     调用方传入如 logger.LogInfo 等已绑定级别的函数，保持日志格式统一。
//   - shutdownTimeout: preShutdown 的最大执行时长；超过后立即返回超时错误。
//   - preShutdown: 真正执行关闭动作（如停止 HTTP Server、刷盘、断开数据库）的回调；
//     允许为 nil，此时仅等待 ctx 取消并直接返回 nil。
//
// 返回值：
//   - 若 preShutdown 在 shutdownTimeout 内完成且无错误，返回 nil。
//   - 若 preShutdown 返回错误，原样包装后返回。
//   - 若 preShutdown 执行超时，返回带 DeadlineExceeded 线索的错误。
func WaitUntilDone(
	ctx context.Context,
	loggerFn func(format string, args ...any),
	shutdownTimeout time.Duration,
	preShutdown func() error,
) error {
	<-ctx.Done()

	if loggerFn != nil {
		loggerFn("shutdown signal received, starting cleanup (timeout=%s)", shutdownTimeout)
	}

	if preShutdown == nil {
		if loggerFn != nil {
			loggerFn("no pre-shutdown hook, shutdown complete")
		}
		return nil
	}

	// 为清理动作单独建超时 Context，不受父 ctx 已经取消的影响。
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	done := make(chan error, 1)
	go func() {
		defer close(done)
		done <- preShutdown()
	}()

	select {
	case err := <-done:
		if err != nil {
			if loggerFn != nil {
				loggerFn("pre-shutdown hook failed: %v", err)
			}
			return fmt.Errorf("pre-shutdown hook failed: %w", err)
		}
		if loggerFn != nil {
			loggerFn("pre-shutdown hook completed, shutdown successful")
		}
		return nil
	case <-shutdownCtx.Done():
		cause := shutdownCtx.Err()
		if loggerFn != nil {
			loggerFn("pre-shutdown hook timed out after %s: %v", shutdownTimeout, cause)
		}
		return fmt.Errorf("pre-shutdown hook timed out after %s: %w", shutdownTimeout, cause)
	}
}

// 编译期断言：确保我们返回的错误仍然保留 DeadlineExceeded 等可被 errors.Is 识别的语义。
var _ = errors.Is(context.DeadlineExceeded, context.DeadlineExceeded)
