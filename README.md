# Argus (观枢)

Kubernetes 原生的 LLM 出站流量安全网关。业务 Pod 零改造，集群级透明引流，对大模型出口请求做 Prompt 提取 / 多检测器流水线 / 风险打分 / 放行或阻断，最后把 AIEvent 归档。

当前是 **阶段一 1-A 完成版**：项目骨架、5 份 gRPC proto、CRD Go 类型、构建脚本、Helm Chart 骨架、README 全在。阶段一 B/C/D（真实业务逻辑、端到端、性能基线）还没写，代码里对应的子包先留了 `doc.go` 占位。

## 做什么 / 不做什么

**做**：
- 集群级透明拦截所有向 LLM 厂商发的请求（Cilium eBPF 或 iptables TPROXY）
- 识别是不是 LLM 流量；非 LLM 流量直接透传，不蹭任何额外延迟
- 对 Prompt 跑 rules / heuristic / encoding / semantic 四段流水线，返回 risk_score
- 按 `ArgusSecurityPolicy` 选 allow / block / degraded，支持 monitor 和 enforce 两种模式
- 把事件流式上报回 controller，落盘成 JSONL（at-least-once）
- 身份解析靠 controller（gateway Pod 自己不抓 k8s 权限）

**不做（阶段一明确不做）**：
- 多厂商 adapter 以外的真实逻辑（只保留 OpenAI 兼容骨架）
- 语义检测器接外部 LLM（预留接口，返回 unimplemented）
- 可视化面板、C++ 高性能检测器、多集群联邦、对象存储 / SIEM 直写

## 目录

```
cmd/
├── argus-gateway/        # 入口：装配依赖 → Run(ctx) → 信号退出
└── argus-controller/     # 同上

internal/
├── utilities/            # slog logger（纯标准库，所有人都能用）
├── config/               # viper + 环境变量 + 默认值
├── detector/
│   ├── rules/            # 正则 + 关键词
│   ├── heuristic/        # 熵 / 字符异常 / 引号不平衡等结构特征
│   ├── encoding/         # Base64 / Hex / Unicode 转义还原后再回灌检测
│   └── semantic/         # 预留接口，真实接入另做
├── gateway/              # 数据平面。禁止依赖 internal/controller
│   ├── server tls proxy protocol prompt adapter pipeline risk policy identity metrics event
├── controller/           # 控制平面。禁止依赖 internal/gateway
│   ├── policy identity health event metrics

pkg/                      # 对外稳定公共库。写之前先在 tasks.md 登记
├── apierrors signal logger tracing

api/argus/v1alpha1        # CRD Go 类型 + GVK + DeepCopy 最小实现
proto/argus/              # 5 份 gRPC 契约
pkg/pb/argus/*/v1alpha1   # protobuf / gRPC 生成桩
deploy/helm/argus         # Helm Chart 骨架（values.yaml + Chart.yaml，模板 placeholder）
scripts/                  # gen-proto.sh / fix-proto-imports.sh
test/{unit,integration,e2e}
```

## 装工具链

Go ≥ 1.22，K8s ≥ 1.27，Helm ≥ 3.10。

```bash
# Go 格式化 / 静态检查
go install golang.org/x/tools/cmd/goimports@latest
go install honnef.co/go/tools/cmd/staticcheck@latest

# Proto
go install github.com/bufbuild/buf/cmd/buf@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# CRD YAML / DeepCopy
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
```

提交前必跑：

```bash
make fmt      # gofmt -s -w .  && goimports -w .
make lint     # go vet ./...   && staticcheck ./...
make test     # go test -race -cover ./...   普通包 ≥70%，detector ≥90%
make build    # bin/argus-gateway + bin/argus-controller
```

## 生成依赖

```bash
make proto   # 产物: pkg/pb/argus/{policy,detection,event,health,identity}/v1alpha1/*.{pb.go,_grpc.pb.go}
make crd     # 产物: deploy/helm/argus/templates/crds/*.yaml  +  api/argus/v1alpha1/zz_generated.deepcopy.go
```

常见坑：IDE 给 `import "argus/detection.proto"` 报红但命令行 `make proto` 是绿的。这是 IDE 的 proto import root 没指到 `proto/` 目录，直接跑：

```bash
bash scripts/fix-proto-imports.sh
```

脚本会跑 `buf build / buf lint / protoc` 四条校验，最后附 VSCode / GoLand 的手动配置步骤。

## 样例

OpenAI Provider + 默认策略（作用于 default 命名空间，semantic 关闭）：

```yaml
# deploy/examples/openai-provider.yaml
apiVersion: argus.cncf/v1alpha1
kind: LLMProvider
metadata:
  name: openai
spec:
  type: openai
  hosts: ["api.openai.com"]
  sni:   ["api.openai.com"]
  upstream:
    base_url: "https://api.openai.com/v1"
    auth_secret_ref:
      name: "openai-api-key"
      key:  "api_key"
  models: ["gpt-4o", "gpt-4o-mini", "gpt-3.5-turbo"]
---
# deploy/examples/default-policy.yaml
apiVersion: argus.cncf/v1alpha1
kind: ArgusSecurityPolicy
metadata:
  name: default-policy
spec:
  providers: [{ name: openai }]
  detectors:
    rules:     { enabled: true }
    heuristic: { enabled: true }
    encoding:  { enabled: true }
    semantic:  { enabled: false }
  thresholds:
    risk: { low: 0.3, medium: 0.6, high: 0.85 }
  scope:
    namespaces: ["default"]
```

```bash
kubectl create secret generic openai-api-key --from-literal=api_key='sk-xxxx'
kubectl apply -f deploy/examples/openai-provider.yaml
kubectl apply -f deploy/examples/default-policy.yaml
kubectl get llmproviders.argus.cncf -o wide
kubectl get argussecuritypolicies.argus.cncf -o yaml | grep -A5 thresholds
```

## 故障排查

**现象 1：`Import 'argus/detection.proto' was not found or had errors`（IDE 红）**

```bash
bash scripts/fix-proto-imports.sh
```

脚本过了 IDE 还红 → 在 IDE 里把 proto include root 设成 `<project>/proto`。

**现象 2：go build 报 `no required module provides package .../pkg/pb/...`**

```bash
make proto && go mod tidy && go build ./...
```

**现象 3：`no kind "ArgusSecurityPolicy" is registered`**

```bash
make crd
helm upgrade argus ./deploy/helm/argus
kubectl api-resources | grep argus
```

**现象 4：gateway 起不来，说 `/tls/tls.crt: no such file or directory`**

开发环境：把 Helm values 里 `gateway.tls.autoGenerate=true` 打开，Chart 自己签一张自签。
生产：`kubectl -n argus get certificates,certificaterequests,orders,challenges` 看 cert-manager 签发到哪一步。

**现象 5：`argus_event_pending_events` 一直涨**

先看 controller 盘满没：
```bash
kubectl -n argus exec deploy/argus-controller -- df -h /var/lib/argus
kubectl -n argus logs deploy/argus-controller -c controller | grep -iE 'event|rpc'
```

事件至少一次投递不丢，堆积通常是 controller 写盘太慢。

**现象 6：rule 命中了但没 block**

看 policy 阈值和真实 detector_score：
```bash
kubectl get argussecuritypolicies default-policy -o yaml | grep -A5 thresholds
# Prometheus: argus_gateway_detector_score{detector="rules"}
```

## Makefile 目标

| 目标 | 说 明 |
|---|---|
| `make all` | fmt → lint → test → build（提 PR 前跑这个） |
| `make help` | 列全部目标 |
| `make fmt lint test build` | 4 个核心门禁 |
| `make proto / crd` | 重新生成桩 / CRD YAML |
| `make tidy / clean` | 整理依赖 / 删 bin pkg/pb |

## 文档索引

- 代码强制规范：`.trae/rules/go-coding-standards.md`
- 产品设计：`.trae/specs/argus-guanshu/spec.md`
- 任务拆解：`.trae/specs/argus-guanshu/tasks.md`
- 验收清单：`.trae/specs/argus-guanshu/checklist.md`

## 测试

```bash
go test -race -cover ./...
```

覆盖率硬门槛（Code Review 不打商量）：普通包 ≥ 70%，`internal/detector/*` ≥ 90%。阶段一 1-A 还没写业务，所以现在是 `[no test files]`，正常。
