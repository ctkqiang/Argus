// Package apierrors 提供 Argus 标准化错误类型、哨兵错误与错误包装工具。
// 所有业务层返回的错误应优先使用本包的 AppError 或哨兵错误进行包装，
// 以便网关层、控制器层统一提取错误码、HTTP 状态码与结构化详情。
package apierrors

import (
	"errors"
	"fmt"
)

// AppError 是 Argus 全链路统一的应用错误类型。
// 携带稳定的机器可读 Code、人类可读 Message、可选的底层 Cause（错误链）
// 以及任意结构化 Details，便于日志记录与 API 响应渲染。
type AppError struct {
	// Code 是稳定的机器可读错误码，大写加下划线风格（如 NOT_FOUND）。
	// 调用方通过 CodeOf 或 IsXxx 系列函数判定错误类型时以 Code 为准。
	Code string

	// Message 是面向终端用户或调用方的人类可读描述，不应包含敏感信息。
	Message string

	// Cause 是触发本错误的底层原因，保留错误链以便调试与 errors.Is/As。
	// 构造 AppError 时允许 Cause 为 nil，代表没有更低层的错误来源。
	Cause error

	// Details 携带任意结构化上下文（如字段校验失败详情、请求ID、策略名）。
	// 键名建议使用 snake_case，与日志字段命名约定保持一致。
	Details map[string]any
}

// Error 实现 error 接口，格式为 `code: message (cause: cause.Error())`。
// 当 Cause 为 nil 时省略 cause 部分，避免输出 `cause: <nil>`。
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (cause: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap 支持 errors.Is / errors.As 沿错误链穿透。
func (e *AppError) Unwrap() error { return e.Cause }

// 哨兵错误定义。哨兵错误用于调用方通过 errors.Is 快速判定类别；
// 具体携带消息与详情的错误请使用 NewXxx 或 Wrap 构造函数。
var (
	// ErrNotFound 代表请求的资源不存在（如策略、提供者、配置项）。
	ErrNotFound = &AppError{Code: "NOT_FOUND", Message: "resource not found"}

	// ErrInvalidInput 代表调用方输入参数校验失败（格式、范围、必填项等）。
	ErrInvalidInput = &AppError{Code: "INVALID_INPUT", Message: "invalid input"}

	// ErrPolicyMissing 代表指定的安全策略不存在或尚未下发。
	ErrPolicyMissing = &AppError{Code: "POLICY_MISSING", Message: "security policy is missing"}

	// ErrProviderNotFound 代表请求的 LLM 提供者未在注册表中注册。
	ErrProviderNotFound = &AppError{Code: "PROVIDER_NOT_FOUND", Message: "LLM provider not found"}

	// ErrDetectorTimeout 代表安全检测器执行超时，未在规定窗口内返回判定结果。
	ErrDetectorTimeout = &AppError{Code: "DETECTOR_TIMEOUT", Message: "security detector timed out"}

	// ErrTLSInspectionFailed 代表 TLS 流量解密/检查阶段失败（证书无效、握手失败等）。
	ErrTLSInspectionFailed = &AppError{Code: "TLS_INSPECTION_FAILED", Message: "TLS inspection failed"}

	// ErrUnauthorized 代表请求未通过身份认证或缺少必要权限。
	ErrUnauthorized = &AppError{Code: "UNAUTHORIZED", Message: "unauthorized"}

	// ErrInternal 代表未分类的服务端内部错误，通常意味着非预期的程序异常。
	ErrInternal = &AppError{Code: "INTERNAL", Message: "internal error"}
)

// NewNotFound 构造一条 NOT_FOUND 错误并携带自定义消息。
func NewNotFound(message string) *AppError {
	return &AppError{Code: ErrNotFound.Code, Message: message}
}

// NewInvalidInput 构造一条 INVALID_INPUT 错误并携带自定义消息。
func NewInvalidInput(message string) *AppError {
	return &AppError{Code: ErrInvalidInput.Code, Message: message}
}

// NewPolicyMissing 构造一条 POLICY_MISSING 错误，并把策略名写入 Details 以便定位。
func NewPolicyMissing(name string) *AppError {
	return &AppError{
		Code:    ErrPolicyMissing.Code,
		Message: fmt.Sprintf("security policy %q is missing", name),
		Details: map[string]any{"policy_name": name},
	}
}

// Wrap 将底层错误包装为指定 code 的 AppError，并附加描述性消息。
// Cause 通过 %w 保留在错误链中，errors.Is/As 可正常穿透。
// 当传入的 code 为空字符串时降级为 INTERNAL，避免未赋值的零值误用。
func Wrap(err error, message string, code string) *AppError {
	if code == "" {
		code = ErrInternal.Code
	}
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   fmt.Errorf("%s: %w", message, err),
	}
}

// IsNotFound 判断错误链中任意一级是否为 NOT_FOUND 类错误。
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Code == ErrNotFound.Code
	}
	return errors.Is(err, ErrNotFound)
}

// IsInvalidInput 判断错误链中任意一级是否为 INVALID_INPUT 类错误。
func IsInvalidInput(err error) bool {
	if err == nil {
		return false
	}
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Code == ErrInvalidInput.Code
	}
	return errors.Is(err, ErrInvalidInput)
}

// CodeOf 提取错误链中的 AppError Code。
// 若错误为 nil 返回空字符串；若错误链中未找到 AppError，为了便于上层统一处理，
// 返回 INTERNAL 以标记"未知/内部"类别；但若是标准库未包装的 nil 仍保持空字符串。
func CodeOf(err error) string {
	if err == nil {
		return ""
	}
	var ae *AppError
	if errors.As(err, &ae) && ae != nil {
		return ae.Code
	}
	return ErrInternal.Code
}
