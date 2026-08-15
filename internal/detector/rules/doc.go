// Package rules 实现基于规则引擎的 Prompt/Response 安全检测，
// 支持 YAML/JSON 格式的自定义规则（正则匹配、关键词、词频统计、相似度匹配）。
// 规则集由控制平面通过 ArgusSecurityPolicy CRD 下发，热加载无需重启。
package rules
