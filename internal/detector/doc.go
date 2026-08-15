// Package detector 定义观枢安全检测引擎的顶层接口与通用类型，
// 所有具体检测器（rules、heuristic、encoding、semantic）必须实现本包定义的 Detector 接口。
// 该包仅依赖 Go 标准库与 internal/utilities，禁止引入 gateway/controller 等上层模块。
package detector
