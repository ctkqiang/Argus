// Package pipeline 编排检测流水线的执行顺序与并发策略，串联协议解析、
// prompt 归一化、多 detector 并行检测、风险聚合、策略决策等阶段。
// 流水线支持超时、短路（高风险直接阻断）与可观测埋点。
package pipeline
