package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LLMProvider 描述一个外部大模型服务提供方的域名、SNI、协议适配等元数据。
// 网关通过匹配 Host/SNI 将出站流量归属于特定 Provider，并据此选择协议解析 Adapter。
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=llmp;llmps
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Adapter",type="string",JSONPath=".spec.adapter"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type LLMProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec 是 LLMProvider 的期望配置，包含域名匹配、上游协议、适配器类型等字段。
	Spec LLMProviderSpec `json:"spec"`

	// Status 是 LLMProvider 的实际状态，占位字段；后续补充探测健康度、域名解析结果等。
	Status LLMProviderStatus `json:"status,omitempty"`
}

// LLMProviderSpec 定义大模型服务提供方的匹配与转发参数，对齐 spec.md §8.2 YAML 字段。
type LLMProviderSpec struct {
	// Type 指定 Provider 的厂商类型，例如 openai-compatible、anthropic、qwen、zhipu 等。
	// 该字段用于事件分类与 UI 展示，不直接驱动协议解析逻辑。
	Type string `json:"type"`

	// Hosts 是匹配该 Provider 的 Host 头域名列表，支持通配符前缀（如 *.openai.com）。
	// 网关在 HTTPS CONNECT 或 HTTP Host 头中按顺序匹配，命中其一即归属本 Provider。
	Hosts []string `json:"hosts"`

	// SNI 是 TLS ClientHello 阶段 Server Name Indication 的匹配列表，语义同 Hosts。
	// 对于纯 HTTPS 流量通常与 Hosts 保持一致；部分厂商存在 SNI 与 Host 不同的场景。
	SNI []string `json:"sni,omitempty"`

	// Upstream 配置上游转发的协议参数。
	Upstream UpstreamSpec `json:"upstream"`

	// Adapter 指定网关解析请求/响应体的协议适配器，例如 openai、anthropic、generic-json。
	// 适配器决定了如何从 HTTP Body 中提取 Prompt、补全、Token 使用量等字段。
	Adapter string `json:"adapter"`

	// Models 是该 Provider 允许或已知的模型 ID 列表，主要用于审计展示与策略细粒度控制。
	// 空列表表示不做模型名维度限制。
	Models []string `json:"models,omitempty"`

	// SkipTLSInspection 控制是否跳过该 Provider 的 TLS 中间人解密。
	// 当为 true 时，网关仅基于 SNI/Host 做 L4 策略匹配，不解析应用层内容，检测器流水线不生效。
	// 适用于使用证书钉扎（Certificate Pinning）无法解密的厂商。
	SkipTLSInspection bool `json:"skipTLSInspection,omitempty"`
}

// LLMProviderStatus 定义 LLMProvider 资源的实际状态，当前为占位结构体。
// 后续将补充 Conditions（健康探测结果）、LastProbeTime、ResolvedHosts 等字段。
type LLMProviderStatus struct {
}

// UpstreamSpec 描述上游转发的协议层参数。
type UpstreamSpec struct {
	// Scheme 指定上游转发时使用的协议方案，目前仅支持 http 与 https。
	// 通常与原请求保持一致，即 https；在内网纯 HTTP 上游场景下可改为 http。
	Scheme string `json:"scheme"`
}

// LLMProviderList 是 LLMProvider 的列表返回类型，用于 K8s LIST API。
//
// +kubebuilder:object:root=true
type LLMProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items 是本次 LIST 响应返回的 LLMProvider 条目集合。
	Items []LLMProvider `json:"items"`
}
