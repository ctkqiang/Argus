# Checklist — Argus（观枢）MVP 阶段一验收

> 对应 `spec.md` §17.1 阶段一交付验收项与 `tasks.md` 全部任务。
> 验收方式：逐项打勾；任何一项未通过均不得声明阶段一完成。
> 验收顺序：先文档与契约 -> 再组件实现 -> 再集成与端到端 -> 最后性能基线。

---

## 1. 硬性架构约束合规性（C1-C5）

- [ ] C1：集群中部署且仅部署一个逻辑 argus-gateway（多副本视为同一逻辑实例）。
- [ ] C2：全集群无任何业务 Pod sidecar、无每容器 Agent、无业务节点级 DaemonSet 传感器（`ds-node-tproxy` 属 argus-system 节点级引流组件，非业务传感器，允许）。
- [ ] C3：业务 Pod 源码 / Dockerfile / LLM SDK 零修改；未在业务侧设置 `OPENAI_BASE_URL` 等代理环境变量（CA 注入属基础设施层，不算业务改造）。
- [ ] C4：流量通过集群全局透明出口网络能力引流，不依赖业务侧代理配置。
- [ ] C5：业务 Pod 直连 LLM 厂商原始域名（如 `api.openai.com`），DNS 解析与连接目标未被业务感知地修改。

## 2. Spec 文档评审（阶段 0）

- [ ] `spec.md` 评审通过，5 条硬性架构约束无歧义。
- [ ] `tasks.md` 评审通过，任务粒度与依赖关系合理。
- [ ] `checklist.md` 评审通过，覆盖 §17.1 全部交付验收项。

## 3. 设计文档交付物（spec §17 输出物）

- [ ] `docs/architecture.md` 包含 Mermaid 逻辑拓扑图与完整报文请求流转路径图。
- [ ] 组件职责矩阵存在且覆盖 §4 全部行（gateway / controller 职责分明）。
- [ ] `docs/traffic-interception.md` 含四方案对比表（Cilium eBPF / Istio Egress / iptables-nftables / VPC 路由），每方案写明优势与硬性限制。
- [ ] `docs/traffic-interception.md` 写明 MVP 首选与兜底方案的安装步骤、前置条件、局限性、回滚步骤。
- [ ] `docs/pod-identity.md` 含溯源链路图与准确性矩阵（普通 / hostNetwork / 已退出 / 多容器 / SNAT）。
- [ ] `docs/tls-design.md` 含两种可见模式（被动观测 / TLS 解密）说明。
- [ ] `docs/tls-design.md` 含 CA 根证书要求与信任模型。
- [ ] `docs/tls-design.md` 含证书生成逻辑与轮换机制。
- [ ] `docs/tls-design.md` 含安全风险影响表。
- [ ] `docs/tls-design.md` 含故障降级行为表。
- [ ] `docs/tls-design.md` 含流量拦截范围与绕过规则；明示禁止不安全 TLS 劫持捷径。
- [ ] `docs/runmodes-failure.md` 含 monitor / enforce 模式说明。
- [ ] `docs/runmodes-failure.md` 含 fail-open / fail-closed 说明。
- [ ] `docs/runmodes-failure.md` 含 §12.3 故障场景矩阵全部 24 格。
- [ ] Go 项目完整目录结构（§15）在仓库中实际存在且与文档一致。

## 4. Protobuf 契约

- [ ] `proto/argus/policy.proto` 定义 `PolicyService` 与 `WatchPolicies` / `WatchProviders`。
- [ ] `proto/argus/detection.proto` 定义 `DetectionService`（`Detect` + `DetectStream`）与 `LLMRequest` / `DetectionContext` / `DetectResponse` / `DetectionResult`。
- [ ] `proto/argus/event.proto` 定义 `EventService.ReportEvents` 与 `AIEvent`，字段编号严格按 spec §14（1-18）。
- [ ] `proto/argus/health.proto` 定义 `HealthService`（`GatewayHeartbeat` + `ControllerHealth`）。
- [ ] `proto/argus/identity.proto` 定义 `PodIdentityService.Lookup`。
- [ ] `make proto` 生成 Go 桩代码且 lint 通过。

## 5. CRD 定义

- [ ] `ArgusSecurityPolicy` CRD 包含 `mode`、`failureMode`、`providers`、`detectors`、`thresholds`、`scope` 字段。
- [ ] `LLMProvider` CRD 包含 `type`、`hosts`、`sni`、`upstream`、`adapter`、`models`、`skipTLSInspection` 字段。
- [ ] `make crd` 生成 CRD YAML，`kubectl apply --dry-run=server` 通过。

## 6. 控制平面 argus-controller

- [ ] 单副本 Deployment + leader election 配置就绪。
- [ ] watch `ArgusSecurityPolicy` / `LLMProvider` 并通过 `PolicyService` 推送全量 + 增量快照。
- [ ] `PodIdentityService.Lookup` 在单元测试中覆盖普通 / hostNetwork / 已退出 / 未命中四类场景。
- [ ] `EventService.ReportEvents` 流式接收 + 本地 PV 持久化（按 `cluster_id/date/event_id.jsonl` 滚动）。
- [ ] `HealthService.ControllerHealth` + `GatewayHeartbeat` 双向健康检查可工作。
- [ ] controller 严禁运行任何检测业务逻辑（代码评审确认）。
- [ ] controller RBAC 仅含必要权限（pods get/list/watch + CRD 全权），无 cluster-admin 越权。

## 7. 数据平面 argus-gateway

- [ ] Deployment 多副本 + HPA 配置就绪，前端挂 Service。
- [ ] 接收透明引流流量，区分 LLM 与非 LLM；非 LLM 透传不进检测流水线。
- [ ] 协议识别支持 TLS / HTTP/1.1 / HTTP/2。
- [ ] LLM 提供商识别基于 SNI / Host 匹配 `LLMProvider` 快照。
- [ ] OpenAI 兼容适配器支持 `/v1/chat/completions` 与 `/v1/completions`，转 `LLMRequest`。
- [ ] Prompt 提取覆盖单轮 / 多轮 / 带 tools 三类。
- [ ] 检测流水线顺序为：规则 -> 启发式 -> 编码混淆 -> 语义（语义接口预留 unimplemented）。
- [ ] 规则检测器覆盖 §13.2 中"直接提示注入 / 指令覆盖 / 系统提示词窃取"基础模式。
- [ ] 编码混淆检测器覆盖 Base64 / Hex / Unicode 转义。
- [ ] 风险打分输出 `risk_score ∈ [0,1]`。
- [ ] 策略引擎支持 monitor / enforce + fail-open / fail-closed 四组合，覆盖 §12.3 矩阵全部格子。
- [ ] SSE 流式代理按 chunk 透传，单连接内存上限 64KiB（可配置）。
- [ ] 背压：慢客户端触发 TCP 反压，不 OOM。
- [ ] `max_connections` 限制 + LRU 关闭可工作。
- [ ] Prometheus 指标端点 `/metrics` 暴露 §5.5 全部 9 个指标。
- [ ] `AIEvent` 字段完整填充（含 `cluster_id` / `session_id` / `request_id` / `request_payload` 截断脱敏）。
- [ ] gateway Pod 不持有任何 K8s RBAC 权限（身份解析全部走 controller）。

## 8. TLS 解密检测

- [ ] B 模式（TLS 解密检测）默认开启，A 模式（被动观测）作为元数据补充可用。
- [ ] 网关根据 ClientHello SNI 动态选择叶子证书完成 TLS 终止。
- [ ] CA 与叶子证书从 `Secret` 加载；缺失时按 `failureMode` 降级。
- [ ] MutatingWebhook 自动向业务 Pod 注入 `ca.crt` + 信任环境变量。
- [ ] 白名单生效：`scope.excludeNamespaces` / `LLMProvider.skipTLSInspection` / Pod 标注 `argus.argus.io/inspect=skip`。
- [ ] 文档与代码均无 `InsecureSkipVerify=true` 或关闭证书校验的捷径。

## 9. 流量引流

- [ ] Cilium eBPF 模板 `CiliumEgressGatewayPolicy` 在装了 Cilium 的集群中可重定向业务流量到网关。
- [ ] iptables/nftables TPROXY 模板在无 Cilium 集群中可透明代理业务流量到网关。
- [ ] 引流层组件均部署在 `argus-system`，业务 Pod 无任何注入。
- [ ] 引流层故障时业务流量直连 LLM（安全降级），并触发告警，不静默丢弃。
- [ ] 文档写明两种方案的前置条件、安装步骤、局限性、回滚步骤。

## 10. Helm Chart 与 K8s 资源

- [ ] `helm lint deploy/helm/argus` 通过。
- [ ] `helm template` 渲染无错，产出 §16 全部资源。
- [ ] 单条 `helm install argus ./deploy/helm/argus` 即可完成部署。
- [ ] values.yaml 暴露：副本数、镜像、资源、运行模式、阈值、引流方案选择。
- [ ] NetworkPolicy 限制网关出口仅至 `LLMProvider.hosts`。
- [ ] 默认策略与 OpenAI Provider 示例可 `kubectl apply` 成功。

## 11. 端到端验证（spec §17.1 第 9 项）

- [ ] kind 测试集群 + Cilium + cert-manager 可一键搭建。
- [ ] 用例 1：未修改业务 Pod -> 正常 Prompt -> 放行 -> SSE 流式响应 -> `AIEvent(action=allow)`。
- [ ] 用例 2：业务 Pod 发"ignore previous instructions" -> enforce 模式阻断 -> `AIEvent(action=block)`。
- [ ] 用例 3：monitor 模式相同 Prompt -> 放行 -> `AIEvent(action=allow, risk_score>0.8)`。
- [ ] 用例 4：fail-closed + 检测器人为超时 -> 阻断 + 降级事件（原因 `detector_timeout`）。
- [ ] 用例 5：业务未挂载 CA -> TLS 握手失败 -> fail-open 透传 / fail-closed 阻断 + 事件。
- [ ] `AIEvent` 落盘文件可被读取，字段完整。
- [ ] 业务 Pod 在整条链路中源码 / Dockerfile / SDK 零修改。

## 12. 性能与资源基线

- [ ] 100 并发 SSE 流持续 5min 压测无 OOM、无连接泄漏。
- [ ] 单连接内存不超过配置上限（64KiB）。
- [ ] 网关总内存随并发有界增长（不超过 `limits.memory`）。
- [ ] 基线报告 `docs/perf-baseline.md` 产出（CPU / 内存 / P99 延迟 / 阻断计数）。

## 13. 安全与运维

- [ ] CA 私钥不落盘到业务可读路径，仅存内存 / Secret。
- [ ] `AIEvent.request_payload` 默认截断 4KiB 并脱敏。
- [ ] 全部故障路径产出 `AIEvent` 或日志，无静默丢弃。
- [ ] 引流层 / 网关 / controller 三层均有就绪与存活探针。
- [ ] 证书过期 Prometheus 告警规则存在。
- [ ] `docs/helm-install.md` 含完整安装 + values 说明 + 引流方案选择 + 回滚步骤。

## 14. 阶段一显式不交付项确认

- [ ] 确认未实现多厂商适配器（仅 OpenAI 兼容）。
- [ ] 确认语义检测器仅接口预留，未接入外部模型。
- [ ] 确认未实现可视化面板。
- [ ] 确认未实现 C++ 高性能检测器。
- [ ] 确认事件持久化仅本地 PV，未接对象存储 / SIEM。
- [ ] 确认未实现多集群联邦。

## 15. 最终交付签字

- [ ] 全部 1-13 项打勾。
- [ ] 第 14 项确认阶段一边界未被悄悄突破。
- [ ] 用户在 `spec.md` 末尾或单独 issue 中签字确认阶段一交付完成。
