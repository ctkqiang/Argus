// Package config 定义 Argus 运行配置模型，并负责从环境变量与配置文件读取配置。
// 当前版本仅提供结构体声明与生产合理的默认值；环境变量与 YAML 文件加载逻辑
// 待后续引入配置解析方案后补全。所有字段命名保持与 Helm values / env 键语义一致，
// 便于未来做零拷贝映射。
package config

// GatewayConfig 描述 argus-gateway 组件的完整运行配置。
// 网关负责接收 Pod 出向流量，执行 LLM 请求解析、策略判定与放行/阻断。
type GatewayConfig struct {
	// ListenAddr 网关对外监听地址，格式为 `host:port`。
	// 空 host 表示绑定所有网卡；生产默认 8443 端口。
	ListenAddr string

	// DefaultMode 未命中任何策略时的默认运行模式。
	// 典型取值："enforce"（按阈值放行/阻断）、"observe"（仅记录不阻断）、"bypass"（直接放行）。
	DefaultMode string

	// TLS 是否启用 TLS 解密能力。启用后需要提供 CA 证书或开启自动签发。
	TLS bool

	// ControllerAddr argus-controller 的 gRPC 服务地址，用于拉取策略与上报事件。
	// Kubernetes 场景下为 Service DNS：`argus-controller.<ns>.svc.cluster.local:8444`。
	ControllerAddr string

	// MaxConnections 单实例最大并发连接数，超出后新连接会被直接拒绝。
	// 用于保护网关进程不被高并发打垮。
	MaxConnections int32

	// StreamBufferBytes 单条流（TCP/HTTP）的缓冲区大小（字节）。
	// 过大会放大内存占用，过小会降低吞吐；默认 64KB 平衡两者。
	StreamBufferBytes int32

	// LogLevel 网关日志最低输出级别，字符串形式：debug / info / warn / error / fatal。
	// 对应 pkg/logger 中 LogLevel 的字面值，便于直接解析。
	LogLevel string

	// LogFormat 日志输出格式："text" 或 "json"。
	LogFormat string

	// MetricsAddr Prometheus 指标暴露地址，格式 `host:port`。
	// 默认与业务监听端口分离，便于运维侧独立采集。
	MetricsAddr string
}

// ControllerConfig 描述 argus-controller 组件的完整运行配置。
// 控制器负责策略下发、事件汇总存储、Leader 选举与 CRD 调和。
type ControllerConfig struct {
	// MetricsAddr Prometheus 指标暴露地址，格式 `host:port`。
	MetricsAddr string

	// EnableLeaderElection 是否启用 Leader 选举。
	// 多副本部署时必须开启，以保证 CRD 调和与事件落盘只有一个实例在执行。
	EnableLeaderElection bool

	// LeaderElectionNamespace 选举资源（Lease）所在的命名空间。
	// 通常等于控制器自身所在命名空间；空字符串时使用 in-cluster 配置推断。
	LeaderElectionNamespace string

	// EventStorageRoot 事件文件存储根目录。
	// Kubernetes 场景下应挂载到 PV 或 EmptyDir；默认 /var/lib/argus/events。
	EventStorageRoot string

	// EventFileMaxBytes 单个事件滚动文件的最大字节数。
	// 超出后自动切分到新文件；默认 100MB，平衡检索效率与文件数量。
	EventFileMaxBytes int64

	// LogLevel 控制器日志最低输出级别，字符串形式。
	LogLevel string

	// LogFormat 控制器日志输出格式："text" 或 "json"。
	LogFormat string

	// GRPCAddr 控制器提供给网关调用的 gRPC 监听地址，格式 `host:port`。
	GRPCAddr string
}

// AppConfig 是 Argus 全局顶层配置容器，聚合网关与控制器两大组件配置。
// 单体二进制模式（如开发调试）下可同时启动两侧，生产建议拆分二进制与配置。
type AppConfig struct {
	// Gateway 网关子配置。
	Gateway GatewayConfig

	// Controller 控制器子配置。
	Controller ControllerConfig

	// ClusterID 当前部署集群的唯一标识，用于事件上报、多集群数据聚合时区分来源。
	// 空字符串时退化为 "unknown-cluster"，但生产建议显式配置。
	ClusterID string
}

// DefaultGateway 返回一份面向生产的网关默认配置。
// 适用于快速启动、样例配置、以及 LoadGatewayFromEnv 中未设置键的回退值。
func DefaultGateway() *GatewayConfig {
	return &GatewayConfig{
		ListenAddr:        ":8443",
		DefaultMode:       "enforce",
		TLS:               true,
		ControllerAddr:    "argus-controller.argus-system.svc.cluster.local:8444",
		MaxConnections:    1024,
		StreamBufferBytes: 65536,
		LogLevel:          "info",
		LogFormat:         "text",
		MetricsAddr:       ":9090",
	}
}

// DefaultController 返回一份面向生产的控制器默认配置。
func DefaultController() *ControllerConfig {
	return &ControllerConfig{
		MetricsAddr:             ":9091",
		EnableLeaderElection:    true,
		LeaderElectionNamespace: "argus-system",
		EventStorageRoot:        "/var/lib/argus/events",
		EventFileMaxBytes:       100 * 1024 * 1024,
		LogLevel:                "info",
		LogFormat:               "text",
		GRPCAddr:                ":8444",
	}
}

// LoadGatewayFromEnv 从环境变量加载 GatewayConfig。
// 本版本作为骨架仅返回默认值；后续将通过环境变量前缀 ARGUS_GATEWAY_* 映射字段。
// 返回错误用于保留签名兼容，未来解析失败时可在此上报。
func LoadGatewayFromEnv() (*GatewayConfig, error) {
	return DefaultGateway(), nil
}

// LoadControllerFromEnv 从环境变量加载 ControllerConfig。
// 与 LoadGatewayFromEnv 同理，当前仅返回默认值。
func LoadControllerFromEnv() (*ControllerConfig, error) {
	return DefaultController(), nil
}
