// Package metrics 定义网关 Prometheus 指标采集器，覆盖 QPS、延迟分位、
// 检测命中率、决策分布、流式 token 吞吐等。指标通过 /metrics 端点暴露，
// 禁止使用任何非 Prometheus 标准的指标库。
package metrics
