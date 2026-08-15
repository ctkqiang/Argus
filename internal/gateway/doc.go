// Package gateway 实现观枢数据平面网关的核心业务逻辑，包含流量接收、协议识别、
// 检测流水线编排、策略决策、SSE 流式代理等功能模块。该包依赖 internal/detector
// 提供的检测能力与 internal/utilities 提供的基础设施，禁止反向依赖。
package gateway
