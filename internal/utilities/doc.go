// Package utilities 提供 Argus 项目共享的基础设施组件，包括结构化日志、错误包装、重试工具等。
// 该包仅允许依赖 Go 标准库，禁止引入任何外部三方库或其他 internal 子包，保持底层纯净，
// 以便网关、控制器等上层模块可以无顾虑地复用。
package utilities
