// Package pipeline 把 protocol / prompt / detector / risk / policy 串起来。
// 顺序：规则 → 启发式 → 编码 → 语义，超时 2s 直接降级不挡业务。
package pipeline
