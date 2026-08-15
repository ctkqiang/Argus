// Package signal 信号到 context 的统一转发：SIGINT / SIGTERM 取消，SIGHUP reload。
// cmd 层就别再自己写 os/signal 处理了，用这个。
package signal
