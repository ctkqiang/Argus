// Package identity 实现 Pod 身份服务控制器，通过 Watch Pod/Endpoint 资源
// 维护 IP → 工作负载/服务账号的映射表，供网关侧 identity 模块查询消费。
// 映射信息通过共享内存或 gRPC 接口下发。
package identity
