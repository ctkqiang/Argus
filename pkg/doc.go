// Package pkg 包含 Argus 项目对外可复用的公共库，下游项目可通过 go get 直接引用。
// 子包保持对 internal/ 零依赖，确保外部使用者不会引入网关/控制器内部实现。
package pkg
