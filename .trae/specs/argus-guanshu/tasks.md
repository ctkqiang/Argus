# Tasks — Argus（观枢）MVP 阶段一

> 本任务列表对应 `spec.md` §17 阶段一范围与边界。
> 评审通过 `spec.md` 后再开始执行本列表。当前阶段不写业务代码，仅产出 spec 文档；脚手架阶段由后续 `开始执行` 指令触发。
> 任务粒度遵循"小而可验证"原则；每个任务都给出验证方式；依赖关系见末尾。

---

## 阶段 0：Spec 评审与基线确认

- [ ] Task 0.1: 完成 `spec.md` 评审，确认 5 条硬性架构约束（C1-C5）无歧义。
  - 验证：用户在 `.trae/specs/argus-guanshu/spec.md` 上明确 sign-off。
- [ ] Task 0.2: 完成 `tasks.md` 评审，确认任务粒度与依赖关系合理。
- [ ] Task 0.3: 完成 `checklist.md` 评审，确认验收项覆盖 §17.1 全部交付验收项。

---

## 阶段 1-A：项目脚手架与 protobuf 契约

- [ ] Task 1.1: 初始化 Go module 与项目目录结构（按 `spec.md` §15）。
  - [ ] SubTask 1.1.1: `go mod init github.com/ctkqiang/argus`，声明 Go 版本（≥ 1.22）。
  - [ ] SubTask 1.1.2: 创建 §15 中除 `internal/` 子模块外的全部目录骨架与 `.gitkeep` / `doc.go` 占位。
  - [ ] SubTask 1.1.3: 添加 `Makefile`（`make build / test / proto / crd / lint / e2e` 目标）。
  - [ ] SubTask 1.1.4: 添加 `.gitignore`、`README.md`（仅项目简介与构建命令，不展开文档）。
  - 验证：`make build` 通过，`tree -L 3` 与 §15 结构一致。

- [ ] Task 1.2: 编写 protobuf 契约（按 `spec.md` §7、§13、§14）。
  - [ ] SubTask 1.2.1: `proto/argus/policy.proto` — `PolicyService`（`WatchPolicies`、`WatchProviders`）+ `PolicySnapshot` / `ProviderSnapshot`。
  - [ ] SubTask 1.2.2: `proto/argus/detection.proto` — `DetectionService`（`Detect`、`DetectStream`）+ `LLMRequest` / `DetectionContext` / `DetectResponse` / `DetectionResult`。
  - [ ] SubTask 1.2.3: `proto/argus/event.proto` — `EventService.ReportEvents` + `AIEvent`（字段编号严格按 §14）+ `ReportAck`。
  - [ ] SubTask 1.2.4: `proto/argus/health.proto` — `HealthService`（`GatewayHeartbeat`、`ControllerHealth`）+ `Heartbeat` / `HealthStatus`。
  - [ ] SubTask 1.2.5: `proto/argus/identity.proto` — `PodIdentityService.Lookup` + `PodIdentityRequest` / `PodIdentityResponse`。
  - [ ] SubTask 1.2.6: `proto/buf.yaml` + `scripts/gen-proto.sh`（基于 `buf generate`）。
  - 验证：`make proto` 生成 Go 桩代码，无 lint 错误。

- [ ] Task 1.3: 配置基础依赖。
  - [ ] SubTask 1.3.1: `go.mod` 引入 `google.golang.org/grpc`、`google.golang.org/protobuf`、`github.com/spf13/viper`、`go.uber.org/zap`、`github.com/prometheus/client_golang`、`k8s.io/api*`、`k8s.io/client-go`、`sigs.k8s.io/controller-runtime`、`k8s.io/apimachinery`。
  - [ ] SubTask 1.3.2: `pkg/logger`、`pkg/signal`、`pkg/apierrors` 通用包实现。
  - 验证：`go mod tidy` 通过，`make lint` 通过。

---

## 阶段 1-B：CRD 与控制平面

- [ ] Task 1.4: 定义 CRD Go 类型（按 `spec.md` §8）。
  - [ ] SubTask 1.4.1: `api/argus/v1alpha1/argussecuritypolicy_types.go`（`Spec`：`mode`、`failureMode`、`providers`、`detectors`、`thresholds`、`scope`）。
  - [ ] SubTask 1.4.2: `api/argus/v1alpha1/llmprovider_types.go`（`Spec`：`type`、`hosts`、`sni`、`upstream`、`adapter`、`models`、`skipTLSInspection`）。
  - [ ] SubTask 1.4.3: `groupversion_info.go` + `zz_generated.deepcopy.go`（`controller-gen` 生成）。
  - [ ] SubTask 1.4.4: `scripts/gen-crd.sh` 生成 `deploy/helm/argus/templates/crds/*.yaml`。
  - 验证：`make crd` 生成 CRD YAML；`kubectl apply --dry-run=server -f templates/crds/` 通过。

- [ ] Task 1.5: 实现 argus-controller 骨架。
  - [ ] SubTask 1.5.1: `cmd/argus-controller/main.go`（`controller-runtime` manager，leader election）。
  - [ ] SubTask 1.5.2: `internal/controller/policy/` — watch `ArgusSecurityPolicy` / `LLMProvider`，缓存为内存快照。
  - [ ] SubTask 1.5.3: 实现 `PolicyService` 服务端，维护 gateway watch 流，推送全量 + 增量快照。
  - [ ] SubTask 1.5.4: 实现 `HealthService.ControllerHealth` + 接收 `GatewayHeartbeat`。
  - [ ] SubTask 1.5.5: `internal/controller/metrics/` — controller 侧 Prometheus 指标。
  - 验证：单元测试覆盖快照 diff 与下发逻辑；`go test ./internal/controller/...` 通过。

- [ ] Task 1.6: 实现 PodIdentityService（按 `spec.md` §10）。
  - [ ] SubTask 1.6.1: `internal/controller/identity/` — 通过 `client-go` informer watch 全集群 Pod，维护 `PodIP -> PodIdentity` 缓存（TTL 默认 5min）。
  - [ ] SubTask 1.6.2: 实现 `PodIdentityService.Lookup`，输入 `(srcIP, srcPort, timestamp)` 返回 Pod 元数据。
  - [ ] SubTask 1.6.3: 处理边界：`hostNetwork` Pod、Pod 已退出、多容器 Pod（MVP 填到 pod 级别）。
  - 验证：单元测试覆盖正常、hostNetwork、已退出、未命中四类场景。

- [ ] Task 1.7: 实现 EventService 服务端与本地持久化。
  - [ ] SubTask 1.7.1: `internal/controller/event/` — 实现 `ReportEvents` 流式接收。
  - [ ] SubTask 1.7.2: 持久化到本地 PV（按 `cluster_id/date/event_id.jsonl` 滚动文件，单文件 100MB 滚动）。
  - [ ] SubTask 1.7.3: 批量 `ReportAck`，记录失败重试。
  - 验证：单元测试覆盖流式接收与文件滚动。

---

## 阶段 1-C：数据平面核心（argus-gateway）

- [ ] Task 1.8: 实现 argus-gateway 骨架与配置加载。
  - [ ] SubTask 1.8.1: `cmd/argus-gateway/main.go`（启动监听端口、gRPC 客户端连接 controller）。
  - [ ] SubTask 1.8.2: `internal/config/` — 从 ConfigMap + Env 读取配置（监听端口、controller 地址、运行模式、阈值）。
  - [ ] SubTask 1.8.3: `internal/gateway/server/` — TCP 监听入口（透明代理端口，默认 8443）。
  - 验证：`make build` 产出 `argus-gateway` 二进制，`./argus-gateway --help` 显示配置项。

- [ ] Task 1.9: 协议识别与 LLM 提供商识别（按 `spec.md` §5.2）。
  - [ ] SubTask 1.9.1: `internal/gateway/protocol/` — 识别 TLS / HTTP/1.1 / HTTP/2（基于 ClientHello + ALPN）。
  - [ ] SubTask 1.9.2: LLM 提供商识别：根据 SNI / Host 匹配 controller 下发的 `LLMProvider` 快照。
  - [ ] SubTask 1.9.3: 非 LLM 流量透传（直接 splice 到上游，不进检测流水线）。
  - 验证：单元测试覆盖 OpenAI / 非 LLM 两类流量分流。

- [ ] Task 1.10: OpenAI 兼容适配器与 Prompt 提取（按 `spec.md` §5.2、§13）。
  - [ ] SubTask 1.10.1: `internal/gateway/adapter/openai/` — 将解密后的请求体解析为 `LLMRequest`（`/v1/chat/completions`、`/v1/completions`）。
  - [ ] SubTask 1.10.2: `internal/gateway/prompt/` — 从 `messages`、`system`、`tools` 提取 Prompt 文本，处理多轮拼接。
  - [ ] SubTask 1.10.3: `internal/gateway/adapter/interface.go` — 适配器接口，预留后续厂商扩展。
  - 验证：单元测试覆盖单轮、多轮、带 tools 三类 OpenAI 请求。

- [ ] Task 1.11: 检测器流水线与风险打分（按 `spec.md` §13.3）。
  - [ ] SubTask 1.11.1: `internal/detector/interface.go` — `Detector` 接口 + `DetectionService` 适配（in-process gRPC，预留远程）。
  - [ ] SubTask 1.11.2: `internal/detector/rules/` — 规则检测器（关键词 + 正则 + 模式库，覆盖 §13.2 中"直接提示注入 / 指令覆盖 / 系统提示词窃取"基础模式）。
  - [ ] SubTask 1.11.3: `internal/detector/heuristic/` — 启发式检测器最小实现（结构特征、长度异常、重复模式）。
  - [ ] SubTask 1.11.4: `internal/detector/encoding/` — 编码混淆检测器最小实现（Base64 / Hex / Unicode 转义解码后跑规则）。
  - [ ] SubTask 1.11.5: `internal/detector/semantic/` — 接口预留，返回 `unimplemented`。
  - [ ] SubTask 1.11.6: `internal/gateway/pipeline/` — 流水线编排，顺序执行四检测器并聚合。
  - [ ] SubTask 1.11.7: `internal/gateway/risk/` — 风险打分（加权聚合，输出 `risk_score ∈ [0,1]`）。
  - 验证：单元测试覆盖各类 Prompt 命中与未命中；流水线顺序与超时处理。

- [ ] Task 1.12: 策略引擎与运行模式（按 `spec.md` §12）。
  - [ ] SubTask 1.12.1: `internal/gateway/policy/` — 接收 `risk_score` + 策略快照，输出 `allow` / `block`。
  - [ ] SubTask 1.12.2: 处理 `monitor` / `enforce`、`fail-open` / `fail-closed` 四种组合。
  - [ ] SubTask 1.12.3: 故障降级：检测器超时、controller 不可达、TLS 解密失败（按 §12.3 矩阵）。
  - 验证：单元测试覆盖 §12.3 故障场景矩阵全部格子。

- [ ] Task 1.13: SSE 流式代理与背压（按 `spec.md` §5.3）。
  - [ ] SubTask 1.13.1: `internal/gateway/proxy/` — 透明代理上游连接，保留原始 Host / SNI。
  - [ ] SubTask 1.13.2: SSE 流式透传：按 chunk 转发，不聚合完整响应。
  - [ ] SubTask 1.13.3: 背压：单连接缓冲上限（默认 64KiB），超出触发 TCP 反压。
  - [ ] SubTask 1.13.4: `max_connections` 限制 + LRU 关闭。
  - 验证：单元 + 集成测试覆盖 SSE 流、慢客户端背压、超限连接拒绝。

- [ ] Task 1.14: 事件上报与指标（按 `spec.md` §5.5、§14）。
  - [ ] SubTask 1.14.1: `internal/gateway/event/` — `EventService.ReportEvents` 流式客户端，失败重试 + 本地落盘兜底。
  - [ ] SubTask 1.14.2: `internal/gateway/identity/` — `PodIdentityService.Lookup` 客户端，缓存 30s。
  - [ ] SubTask 1.14.3: `internal/gateway/metrics/` — 暴露 §5.5 全部 Prometheus 指标。
  - [ ] SubTask 1.14.4: `AIEvent` 字段填充（含 `cluster_id` 计算、`session_id` 生成、`request_payload` 截断脱敏）。
  - 验证：`/metrics` 端点暴露全部指标；事件字段填充单元测试。

---

## 阶段 1-D：TLS 解密检测

- [ ] Task 1.15: TLS 解密检测实现（按 `spec.md` §11）。
  - [ ] SubTask 1.15.1: `internal/gateway/tls/` — 基于 ClientHello SNI 动态选择叶子证书完成 TLS 终止。
  - [ ] SubTask 1.15.2: 证书加载：从 `Secret/argus-ca-tls` 加载 CA，从 `Secret/argus-leaf-tls` 加载多域名叶子证书。
  - [ ] SubTask 1.15.3: 被动观测模式（A 模式）：仅采集 SNI / ALPN / 字节数，不解密。
  - [ ] SubTask 1.15.4: TLS 解密失败降级：按 `failureMode` 处理。
  - [ ] SubTask 1.15.5: 白名单与绕过：`scope.excludeNamespaces`、`LLMProvider.skipTLSInspection`、Pod 标注 `argus.argus.io/inspect=skip`。
  - 验证：集成测试覆盖解密成功 / 证书缺失 / 白名单三类场景。

- [ ] Task 1.16: MutatingWebhook 实现 CA 注入。
  - [ ] SubTask 1.16.1: `cmd/argus-controller/` 内增加 webhook server，监听 Pod CREATE。
  - [ ] SubTask 1.16.2: 注入 `ca.crt` 到 `/etc/ssl/certs/argus-ca.pem` + 环境变量 `SSL_CERT_FILE` / `REQUESTS_CA_BUNDLE` / `NODE_EXTRA_CA_CERTS`。
  - [ ] SubTask 1.16.3: `templates/mutatingwebhook.yaml` + webhook 证书 Secret。
  - [ ] SubTask 1.16.4: 失败处理：注入失败时按 `failureMode` 决定是否阻断 Pod 启动（默认不阻断，仅告警）。
  - 验证：kind 集群测试业务 Pod 启动后 `ca.crt` 正确挂载。

---

## 阶段 1-E：流量引流方案

- [ ] Task 1.17: 实现 Cilium eBPF 引流模板（按 `spec.md` §9.2.1）。
  - [ ] SubTask 1.17.1: `deploy/helm/argus/templates/cilium/egress-policy.yaml` — `CiliumEgressGatewayPolicy` 匹配业务 Pod -> argus-gateway。
  - [ ] SubTask 1.17.2: `docs/traffic-interception.md` 写明 Cilium 前置条件（内核版本、CNI、DNS proxy）与安装步骤。
  - [ ] SubTask 1.17.3: 局限性声明：托管 K8s、动态 IP 等。
  - 验证：安装了 Cilium 的 kind 集群中业务 Pod 直连 `api.openai.com` 被重定向到网关。

- [ ] Task 1.18: 实现 iptables/nftables TPROXY 兜底模板（按 `spec.md` §9.2.3）。
  - [ ] SubTask 1.18.1: `deploy/helm/argus/templates/iptables/ds-node-tproxy.yaml` — 节点级 DaemonSet（hostNetwork），运行 TPROXY 规则脚本。
  - [ ] SubTask 1.18.2: 配套 IP 列表同步：定期解析 `LLMProvider.hosts` 得到 IP 集合，更新 nftables set。
  - [ ] SubTask 1.18.3: 多副本网关 DNAT 分流（基于 consistent hash）。
  - [ ] SubTask 1.18.4: 文档：局限性（无 SNI 精准匹配、IP 动态变化）、与 kube-proxy 共存注意事项。
  - 验证：无 Cilium 的 kind 集群中业务 Pod 流量被透明代理到网关。

---

## 阶段 1-F：Helm Chart 与 K8s 资源清单

- [ ] Task 1.19: 完成 Helm Chart 骨架（按 `spec.md` §16）。
  - [ ] SubTask 1.19.1: `deploy/helm/argus/Chart.yaml`（`apiVersion: v2`，`name: argus`，`version: 0.1.0`）。
  - [ ] SubTask 1.19.2: `values.yaml` — gateway/controller 副本数、镜像、资源、运行模式、阈值、引流方案选择。
  - [ ] SubTask 1.19.3: `templates/gateway-deployment.yaml` + HPA。
  - [ ] SubTask 1.19.4: `templates/gateway-service.yaml` + `templates/controller-service.yaml`。
  - [ ] SubTask 1.19.5: `templates/controller-deployment.yaml`（leader election）。
  - [ ] SubTask 1.19.6: `templates/configmap.yaml` + `templates/secret-ca.yaml`（CA 生成 Job）。
  - [ ] SubTask 1.19.7: `templates/rbac.yaml`（controller: pods get/list/watch, CRD 全权；gateway: 无 K8s 权限）。
  - [ ] SubTask 1.19.8: `templates/networkpolicy.yaml`（限制网关出口仅至 LLMProvider.hosts）。
  - [ ] SubTask 1.19.9: `templates/crds/argussecuritypolicy.yaml` + `templates/crds/llmprovider.yaml`。
  - [ ] SubTask 1.19.10: `templates/tests/` — helm test（基本连通性）。
  - 验证：`helm lint deploy/helm/argus` 通过；`helm template` 渲染无错。

- [ ] Task 1.20: 默认策略与 Provider 示例。
  - [ ] SubTask 1.20.1: `deploy/examples/policy-default.yaml` — 默认 enforce + fail-open 策略。
  - [ ] SubTask 1.20.2: `deploy/examples/provider-openai.yaml` + `provider-azure.yaml`。
  - 验证：`kubectl apply -f deploy/examples/` 成功。

---

## 阶段 1-G：文档

- [ ] Task 1.21: `docs/architecture.md` — 嵌入 `spec.md` §3 Mermaid 图 + 文字说明。
- [ ] Task 1.22: `docs/traffic-interception.md` — 引流方案对比表 + Cilium/iptables 详细安装步骤 + 局限性 + 回滚步骤。
- [ ] Task 1.23: `docs/pod-identity.md` — §10 溯源设计 + 准确性矩阵 + 边界场景。
- [ ] Task 1.24: `docs/tls-design.md` — §11 完整 TLS 设计 + 信任模型 + 风险矩阵 + 故障降级。
- [ ] Task 1.25: `docs/runmodes-failure.md` — §12 运行模式 + §12.3 故障场景矩阵。
- [ ] Task 1.26: `docs/helm-install.md` — `helm install` 完整流程 + values 说明 + 引流方案选择。
- [ ] Task 1.27: `docs/e2e-verification.md` — 端到端验证步骤与预期结果。

---

## 阶段 1-H：端到端验证

- [ ] Task 1.28: 搭建 kind 测试集群与 Cilium。
  - [ ] SubTask 1.28.1: `test/e2e/kind-config.yaml`。
  - [ ] SubTask 1.28.2: `scripts/e2e/setup-cluster.sh`（kind + Cilium + cert-manager）。
  - 验证：`./scripts/e2e/setup-cluster.sh` 成功，集群就绪。

- [ ] Task 1.29: 端到端用例：未修改业务 Pod -> argus-gateway -> OpenAI 兼容 LLM。
  - [ ] SubTask 1.29.1: 部署一个 mock OpenAI 兼容上游（in-cluster，返回 SSE 流）。
  - [ ] SubTask 1.29.2: 部署业务 Pod（curl + python openai SDK），不修改任何配置。
  - [ ] SubTask 1.29.3: 用例 1：正常 Prompt -> 放行 -> SSE 流式响应 -> `AIEvent(action=allow)`。
  - [ ] SubTask 1.29.4: 用例 2：包含 "ignore previous instructions" 的 Prompt -> enforce 模式下阻断 -> `AIEvent(action=block)`。
  - [ ] SubTask 1.29.5: 用例 3：monitor 模式下相同 Prompt -> 放行 -> `AIEvent(action=allow, risk_score>0.8)`。
  - [ ] SubTask 1.29.6: 用例 4：fail-closed + 检测器人为超时 -> 阻断 + 降级事件。
  - [ ] SubTask 1.29.7: 用例 5：业务未挂载 CA（关闭 webhook）-> TLS 握手失败 + fail-open 透传 / fail-closed 阻断。
  - 验证：`make e2e` 全部用例通过；`AIEvent` 落盘文件可被读取校验。

- [ ] Task 1.30: 性能与资源基线测试。
  - [ ] SubTask 1.30.1: 压测脚本：100 并发 SSE 流，持续 5min。
  - [ ] SubTask 1.30.2: 记录网关 CPU / 内存 / P99 延迟 / 阻断计数。
  - [ ] SubTask 1.30.3: 验证单连接内存不超过 64KiB + 总内存有界。
  - 验证：产出基线报告 `docs/perf-baseline.md`（阶段一仅记录，不设硬指标）。

---

## Task Dependencies

- Task 0.*（评审）阻塞所有后续任务。
- Task 1.2（protobuf）阻塞 Task 1.5、1.6、1.7、1.14。
- Task 1.3（依赖）阻塞 Task 1.4 之后所有 Go 实现。
- Task 1.4（CRD 类型）阻塞 Task 1.5、1.19。
- Task 1.5（controller 骨架）阻塞 Task 1.6、1.7、1.16。
- Task 1.8（gateway 骨架）阻塞 Task 1.9-1.14、1.15。
- Task 1.11（检测器）阻塞 Task 1.12（策略引擎依赖风险分）、Task 1.29（e2e 依赖检测器）。
- Task 1.15（TLS）阻塞 Task 1.16（webhook 依赖证书机制）、Task 1.29 用例 5。
- Task 1.17、1.18（引流）阻塞 Task 1.29（e2e 依赖引流）。
- Task 1.19（Helm）阻塞 Task 1.28（e2e 集群部署）。
- Task 1.21-1.27（文档）可与 Task 1.8-1.18 **并行**（文档基于 spec，不依赖代码完成）。
- Task 1.28-1.30（e2e）依赖 Task 1.8-1.20 全部完成。

### 可并行任务分组

| 并行组 | 任务 |
| --- | --- |
| G1（脚手架后） | 1.4 / 1.21-1.27 文档 |
| G2（controller 骨架后） | 1.6 / 1.7 / 1.16 |
| G3（gateway 骨架后） | 1.9 / 1.10 / 1.13 / 1.14 / 1.15 |
| G4（检测器后） | 1.11 -> 1.12 -> 1.29 |
| G5（引流层） | 1.17 / 1.18 与 G3 并行 |
