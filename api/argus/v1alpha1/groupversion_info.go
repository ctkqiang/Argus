package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// SchemeGroupVersion 是 Argus CRD API 组与版本的唯一标识。
// Group 固定为 argus.argus.io，当前版本为 v1alpha1。
var SchemeGroupVersion = schema.GroupVersion{
	Group:   "argus.argus.io",
	Version: "v1alpha1",
}

// SchemeBuilder 用于批量注册本版本内的所有资源类型到 runtime.Scheme。
// controller-gen 会基于此生成 deepcopy 与 CRD 清单。
var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// Scheme 是本版本独立的 runtime.Scheme 实例，用于本地编码/解码测试。
var Scheme = runtime.NewScheme()

// Codecs 是基于 Scheme 构建的编解码器工厂，支持 JSON 与 YAML。
var Codecs = serializer.NewCodecFactory(Scheme)

// Resource 根据复数资源名构造带 GroupVersion 的 GroupVersionResource。
// 例如 Resource("argussecuritypolicies") → argus.argus.io/v1alpha1, Resource=argussecuritypolicies。
func Resource(resource string) schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(resource)
}

// Kind 根据类型名构造带 GroupVersion 的 GroupVersionKind。
// 例如 Kind("ArgusSecurityPolicy") → argus.argus.io/v1alpha1, Kind=ArgusSecurityPolicy。
func Kind(kind string) schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind(kind)
}

// addKnownTypes 向给定 Scheme 注册本 API 版本包含的所有已知对象类型。
// 同时注册 v1.ListOptions 与 v1.GetOptions，以支持标准 K8s LIST/GET 子资源语义。
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&ArgusSecurityPolicy{},
		&ArgusSecurityPolicyList{},
		&LLMProvider{},
		&LLMProviderList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

func init() {
	if err := AddToScheme(Scheme); err != nil {
		panic("failed to add argus v1alpha1 types to scheme: " + err.Error())
	}
}
