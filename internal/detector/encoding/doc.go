// Package encoding 实现编码混淆检测器，负责识别并解码 Base64、URL 编码、
// Unicode 逃逸、Hex、Rot13 以及多层嵌套编码，还原后交给上层 rules/heuristic
// 做进一步语义检测。检测结果包含原始 payload 与解码链。
package encoding
