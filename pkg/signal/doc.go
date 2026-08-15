// Package signal 封装 POSIX 信号监听与优雅关停逻辑，
// 统一处理 SIGINT、SIGTERM、SIGHUP 并转发给 context.WithCancel，
// 避免 cmd 层与各组件重复实现信号处理。
package signal
