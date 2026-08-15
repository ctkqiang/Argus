// Package config 读 gateway 和 controller 共用的配置：YAML + env + 命令行，后者覆盖前者。
// token / 密钥这类字段只能走 env 或 k8s Secret，别写进配置文件也别打日志。
package config
