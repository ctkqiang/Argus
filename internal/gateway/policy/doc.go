// Package policy 实现策略决策点（PDP），根据 ArgusSecurityPolicy CRD 下发的规则
// 与 risk 层的聚合评分，做出 allow / block / redirect / mask 四类动作决策。
// 支持规则热更新与默认兜底策略。
package policy
