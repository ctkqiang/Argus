// Package encoding 处理编码绕过：Base64 / URL / Unicode / Hex 嵌套解码，
// 解出来的明文再交回给 rules / heuristic 重跑一遍。
package encoding
