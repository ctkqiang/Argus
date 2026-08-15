package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ArgusSecurityPolicy 是集群级安全策略 CRD，定义大模型出口流量的检测规则、
// 阻断阈值与作用域。该资源由 argus-controller 消费并下发到 argus-gateway 数据面。
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=asp;asps
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Mode",type="string",JSONPath=".spec.mode"
// +kubebuilder:printcolumn:name="FailureMode",type="string",JSONPath=".spec.failureMode"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type ArgusSecurityPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec 是 ArgusSecurityPolicy 的期望状态，包含检测模式、检测器配置、阈值与作用域。
	Spec ArgusSecurityPolicySpec `json:"spec"`

	// Status 是 ArgusSecurityPolicy 的实际状态，由控制器回写下发结果与健康信息。
	// 当前阶段为占位结构体，后续迭代中补充条件、同步时间戳等字段。
	Status ArgusSecurityPolicyStatus `json:"status,omitempty"`
}

// ArgusSecurityPolicySpec 定义安全策略的期望配置项，对齐 spec.md §8.1 YAML 字段。
type ArgusSecurityPolicySpec struct {
	// Mode 指定策略运行模式。
	// 枚举值：monitor（仅检测不阻断，只产出日志与事件）、enforce（超过阈值时阻断请求）。
	Mode string `json:"mode"`

	// FailureMode 指定检测器/网关异常时的故障处理模式。
	// 枚举值：fail-open（异常时放行，可用性优先）、fail-closed（异常时阻断，安全性优先）。
	FailureMode string `json:"failureMode"`

	// Providers 指定本策略绑定的 LLMProvider 名称列表，仅对匹配的 Provider 流量生效。
	Providers []ProviderRef `json:"providers,omitempty"`

	// Detectors 配置四类检测器的开关与参数。
	Detectors DetectorConfigSpec `json:"detectors"`

	// Thresholds 配置风险打分的阻断与日志阈值，取值范围 [0.0, 1.0]。
	Thresholds ThresholdSpec `json:"thresholds"`

	// Scope 指定策略生效的命名空间与工作负载范围。
	Scope ScopeSpec `json:"scope"`
}

// ArgusSecurityPolicyStatus 定义安全策略的实际状态，占位字段。
// 后续将补充 Conditions（下发成功/失败）、ObservedGeneration、GatewayRefs 等信息。
type ArgusSecurityPolicyStatus struct {
}

// ProviderRef 引用一个同命名空间（或集群级）下的 LLMProvider 资源名称。
type ProviderRef struct {
	// Name 是被引用 LLMProvider 资源的 metadata.name。
	Name string `json:"name"`
}

// DetectorConfigSpec 聚合四类安全检测器的配置。
type DetectorConfigSpec struct {
	// Rules 配置基于正则/关键字规则的检测器参数。
	Rules RuleDetectorSpec `json:"rules"`

	// Heuristic 配置启发式检测器（如异常分隔符、可疑指令词频）的开关。
	Heuristic ToggleSpec `json:"heuristic"`

	// Encoding 配置编码变形检测器（如 Base64 注入、Unicode 混淆）的开关。
	Encoding ToggleSpec `json:"encoding"`

	// Semantic 配置语义相似度检测器的开关；MVP 阶段默认关闭，依赖外部嵌入模型。
	Semantic ToggleSpec `json:"semantic"`
}

// RuleDetectorSpec 配置基于规则的检测器，包含规则条目列表与总开关。
type RuleDetectorSpec struct {
	// Enabled 控制规则检测器是否启用。
	Enabled bool `json:"enabled"`

	// Rules 是自定义规则项列表，每条规则包含匹配模式与权重。
	Rules []RuleSpecItem `json:"rules,omitempty"`
}

// RuleSpecItem 描述单条检测规则的内容与权重。
type RuleSpecItem struct {
	// ID 是规则的唯一标识符，用于事件溯源与规则命中统计，建议采用 stable-id 风格。
	ID string `json:"id"`

	// Pattern 是规则匹配模式，具体语法由 Type 字段决定（如 regex、keyword）。
	Pattern string `json:"pattern"`

	// Type 指定 Pattern 的匹配类型，例如 regex、keyword、glob 等。
	Type string `json:"type"`

	// Weight 是该规则命中时贡献的风险分数权重，取值建议在 [0.0, 1.0] 区间。
	Weight float64 `json:"weight"`
}

// ToggleSpec 提供统一的检测器开关配置，仅包含 Enabled 字段。
type ToggleSpec struct {
	// Enabled 控制对应检测器是否启用。
	Enabled bool `json:"enabled"`
}

// ThresholdSpec 定义风险打分的两级阈值：日志阈值与阻断阈值。
type ThresholdSpec struct {
	// BlockScore 是阻断阈值；综合风险分 ≥ 该值且 Mode=enforce 时，网关拒绝请求。
	// 取值范围 [0.0, 1.0]，通常应大于 LogScore。
	BlockScore float64 `json:"blockScore"`

	// LogScore 是日志阈值；综合风险分 ≥ 该值时，无论 Mode 为何，均产出安全事件与日志。
	// 取值范围 [0.0, 1.0]。
	LogScore float64 `json:"logScore"`
}

// ScopeSpec 定义策略生效的命名空间与工作负载匹配范围。
type ScopeSpec struct {
	// Namespaces 是策略生效的命名空间列表；使用 ["*"] 表示全部命名空间。
	// 不支持 glob 通配符，仅精确匹配与 "*" 全匹配两种形式。
	Namespaces []string `json:"namespaces"`

	// Workloads 是策略生效的工作负载粒度匹配列表；空列表表示命名空间内全部工作负载。
	Workloads []WorkloadMatchSpec `json:"workloads,omitempty"`
}

// WorkloadMatchSpec 按 Kubernetes 工作负载 Kind 与 Name 进行精确匹配。
type WorkloadMatchSpec struct {
	// Kind 指定工作负载类型，例如 Deployment、StatefulSet、DaemonSet、Job 等。
	Kind string `json:"kind"`

	// Name 指定工作负载的 metadata.name；当前仅支持精确匹配，后续迭代可扩展为选择器。
	Name string `json:"name"`
}

// ArgusSecurityPolicyList 是 ArgusSecurityPolicy 的列表返回类型，用于 K8s LIST API。
//
// +kubebuilder:object:root=true
type ArgusSecurityPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items 是本次 LIST 响应返回的 ArgusSecurityPolicy 条目集合。
	Items []ArgusSecurityPolicy `json:"items"`
}
