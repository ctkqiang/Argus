// Package config 提供网关与控制平面共享的配置加载逻辑，支持 YAML 配置文件、
// 环境变量、命令行参数三层合并。敏感字段（token、密钥）仅允许通过 env 或 k8s Secret
// 注入，禁止出现在配置文件或日志中。该包仅依赖 Go 标准库，禁止引入其他 internal 子包。
package config
