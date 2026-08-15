// Package adapter 定义上游 LLM Provider 适配层接口，各厂商实现放在子目录中。
// 适配器负责将归一化的内部请求转换为厂商特定格式，并将响应反向转换。
package adapter
