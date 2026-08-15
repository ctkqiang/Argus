package v1alpha1

// 本文件提供 CRD 类型的最小 DeepCopy 实现，用于满足 runtime.Object 接口约束
// 以保证脚手架阶段可以编译通过。正式项目应使用 controller-gen 生成
// zz_generated.deepcopy.go 替换本文件。保持所有子结构的切片字段按值深拷贝。

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto 将接收者按值深拷贝写入 out。
func (in *ArgusSecurityPolicy) DeepCopyInto(out *ArgusSecurityPolicy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy 返回接收者的全新深拷贝对象。
func (in *ArgusSecurityPolicy) DeepCopy() *ArgusSecurityPolicy {
	if in == nil {
		return nil
	}
	out := new(ArgusSecurityPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject 满足 runtime.Object 接口，返回深拷贝后的 runtime.Object。
func (in *ArgusSecurityPolicy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto 深拷贝列表类型。
func (in *ArgusSecurityPolicyList) DeepCopyInto(out *ArgusSecurityPolicyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]ArgusSecurityPolicy, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

// DeepCopy 返回 ArgusSecurityPolicyList 的深拷贝。
func (in *ArgusSecurityPolicyList) DeepCopy() *ArgusSecurityPolicyList {
	if in == nil {
		return nil
	}
	out := new(ArgusSecurityPolicyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject 满足 runtime.Object 接口。
func (in *ArgusSecurityPolicyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto 深拷贝策略 Spec。
func (in *ArgusSecurityPolicySpec) DeepCopyInto(out *ArgusSecurityPolicySpec) {
	*out = *in
	if in.Providers != nil {
		out.Providers = make([]ProviderRef, len(in.Providers))
		copy(out.Providers, in.Providers)
	}
	in.Detectors.DeepCopyInto(&out.Detectors)
	out.Thresholds = in.Thresholds
	in.Scope.DeepCopyInto(&out.Scope)
}

// DeepCopy 返回 ArgusSecurityPolicySpec 的深拷贝。
func (in *ArgusSecurityPolicySpec) DeepCopy() *ArgusSecurityPolicySpec {
	if in == nil {
		return nil
	}
	out := new(ArgusSecurityPolicySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto 深拷贝检测器配置。
func (in *DetectorConfigSpec) DeepCopyInto(out *DetectorConfigSpec) {
	*out = *in
	in.Rules.DeepCopyInto(&out.Rules)
	out.Heuristic = in.Heuristic
	out.Encoding = in.Encoding
	out.Semantic = in.Semantic
}

// DeepCopy 返回 DetectorConfigSpec 的深拷贝。
func (in *DetectorConfigSpec) DeepCopy() *DetectorConfigSpec {
	if in == nil {
		return nil
	}
	out := new(DetectorConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto 深拷贝规则检测器配置。
func (in *RuleDetectorSpec) DeepCopyInto(out *RuleDetectorSpec) {
	*out = *in
	if in.Rules != nil {
		out.Rules = make([]RuleSpecItem, len(in.Rules))
		copy(out.Rules, in.Rules)
	}
}

// DeepCopy 返回 RuleDetectorSpec 的深拷贝。
func (in *RuleDetectorSpec) DeepCopy() *RuleDetectorSpec {
	if in == nil {
		return nil
	}
	out := new(RuleDetectorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto 深拷贝作用域。
func (in *ScopeSpec) DeepCopyInto(out *ScopeSpec) {
	*out = *in
	if in.Namespaces != nil {
		out.Namespaces = make([]string, len(in.Namespaces))
		copy(out.Namespaces, in.Namespaces)
	}
	if in.Workloads != nil {
		out.Workloads = make([]WorkloadMatchSpec, len(in.Workloads))
		copy(out.Workloads, in.Workloads)
	}
}

// DeepCopy 返回 ScopeSpec 的深拷贝。
func (in *ScopeSpec) DeepCopy() *ScopeSpec {
	if in == nil {
		return nil
	}
	out := new(ScopeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto 深拷贝 LLMProvider 顶层类型。
func (in *LLMProvider) DeepCopyInto(out *LLMProvider) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy 返回 LLMProvider 的深拷贝。
func (in *LLMProvider) DeepCopy() *LLMProvider {
	if in == nil {
		return nil
	}
	out := new(LLMProvider)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject 满足 runtime.Object 接口。
func (in *LLMProvider) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto 深拷贝 LLMProviderList。
func (in *LLMProviderList) DeepCopyInto(out *LLMProviderList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]LLMProvider, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

// DeepCopy 返回 LLMProviderList 的深拷贝。
func (in *LLMProviderList) DeepCopy() *LLMProviderList {
	if in == nil {
		return nil
	}
	out := new(LLMProviderList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject 满足 runtime.Object 接口。
func (in *LLMProviderList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto 深拷贝 LLMProviderSpec。
func (in *LLMProviderSpec) DeepCopyInto(out *LLMProviderSpec) {
	*out = *in
	if in.Hosts != nil {
		out.Hosts = make([]string, len(in.Hosts))
		copy(out.Hosts, in.Hosts)
	}
	if in.SNI != nil {
		out.SNI = make([]string, len(in.SNI))
		copy(out.SNI, in.SNI)
	}
	out.Upstream = in.Upstream
	if in.Models != nil {
		out.Models = make([]string, len(in.Models))
		copy(out.Models, in.Models)
	}
}

// DeepCopy 返回 LLMProviderSpec 的深拷贝。
func (in *LLMProviderSpec) DeepCopy() *LLMProviderSpec {
	if in == nil {
		return nil
	}
	out := new(LLMProviderSpec)
	in.DeepCopyInto(out)
	return out
}
