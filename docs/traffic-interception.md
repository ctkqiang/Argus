# 流量引流方案设计

> 文档说明：本文档对应 `spec.md` §9，详述 Argus 四种透明出口引流方案的对比、Cilium eBPF 与 iptables/nftables TPROXY 两种 MVP 方案的安装步骤、前置条件、验证、局限性、回滚，以及 Istio 经典 sidecar 不可选的原因与 MVP 最终选型结论。本文档为 [architecture.md](./architecture.md) 的流量引流层补充。

## 1. 概述

Kubernetes `Service` 自身无法捕获任意出站 HTTPS 流量：业务 Pod 直连 `api.openai.com` 的请求会被 kube-proxy 直接 DNAT 到节点路由，绕过任何应用层网关。要实现"业务零改造 + 集群级单逻辑网关"的架构，必须依赖集群全局透明出口网络能力将匹配流量重定向到 `argus-gateway`。

Argus 设计上支持四种引流方案（Cilium eBPF / Istio Egress / iptables-nftables / VPC 路由），MVP 阶段选定 **Cilium eBPF 为首选方案、iptables/nftables TPROXY 为兜底方案**，两者二选一即可运行。**任何方案均不退回 Sidecar 架构**，详见 spec §2 中 C2 / C4 约束。

## 2. 四方案对比表

下表横向对比四种方案在透明劫持能力、业务改造、内核依赖、性能损耗、运维复杂度与硬性限制上的差异。

| 方案 | 透明劫持能力 | 业务改造 | 内核依赖 | 性能损耗 | 运维复杂度 | 主要硬性限制 |
| --- | :---: | :---: | --- | :---: | :---: | --- |
| **Cilium eBPF** | 强（按 Pod/域名/SNI 重定向） | 无 | 内核 ≥ 5.4，Cilium ≥ 1.11 | 低 | 中 | 需 Cilium 作为 CNI；老内核不支持；与某些云厂商 CNI 互斥 |
| **Istio Egress Gateway** | 中（依赖 ServiceEntry + Sidecar 注入或 ambient mode） | **ambient 模式下零改造；经典 sidecar 模式违背 C2** | 无 | 中 | 高 | 经典 Istio 必须 sidecar 注入，**违背 C2 不可选**；ambient 模式成熟度待评估 |
| **iptables / nftables TPROXY** | 中（按目的 IP/端口重定向，需配合 DNS 解析） | 无 | 通用 Linux | 中 | 中 | 无法基于 SNI 精准匹配，需 DNS 解析前置；多副本网关需 DNAT 分流 |
| **云厂商 VPC 路由** | 弱（按 CIDR 路由到网关 ENI） | 无 | 无 | 低 | 低 | 仅按 IP 路由，无法基于域名；LLM 厂商 IP 动态变化；跨可用区流量费用 |

四方案在 Argus 中的角色定位：

- **Cilium eBPF**：MVP 首选，性能最优、与 NetworkPolicy 自然融合。
- **iptables / nftables TPROXY**：MVP 兜底，通用性最强、无 CNI 绑定。
- **Istio Egress Gateway**：仅 ambient 模式可考虑，MVP 不作为首选。
- **云厂商 VPC 路由**：仅作为兜底兜底，不建议 MVP 使用。

## 3. Cilium eBPF 详细方案（MVP 首选）

### 3.1 方案原理

通过 Cilium `CiliumNetworkPolicy` + `CiliumEgressGatewayPolicy`，或基于 eBPF `bpf_sk_assign` / `bpf_msg_redirect` 在 socket 层将匹配目的的连接重定向到 `argus-gateway`。

匹配维度：

- Pod identity（Cilium identity）
- 目的 FQDN（Cilium DNS proxy / SNI）
- 目的 IP/CIDR
- 目的端口

### 3.2 优势

- 真正透明，Pod 无感知；不修改业务 DNS 与连接目标。
- 性能损耗低，eBPF 在内核态完成重定向，无需用户态切换。
- 与 Cilium NetworkPolicy 自然融合，便于落地流量拦截范围与白名单。

### 3.3 前置条件

| 前置项 | 要求 | 验证命令 |
| --- | --- | --- |
| 内核版本 | ≥ 5.4（建议 ≥ 5.10） | `uname -r` |
| CNI | Cilium ≥ 1.11，作为集群唯一 CNI | `cilium status --wait` |
| Cilium DNS proxy | 必须启用（FQDN 匹配依赖） | `cilium config view \| grep toFQDNs` |
| Kube-proxy 模式 | 推荐 Cilium 替代 kube-proxy（kube-proxy replacement） | `cilium status \| grep "Kube-Proxy"` |
| 托管 K8s | 不支持 GKE Autopilot / EKS Fargate 等限制 eBPF 的托管方案 | 文档声明 |
| LLM 厂商 IP 动态性 | 依赖 FQDN 模式，DNS proxy 必须长期运行 | `cilium endpoint list` |

### 3.4 安装步骤

#### 3.4.1 安装 Cilium（如未安装）

```bash
# 添加 Cilium Helm 仓库
helm repo add cilium https://helm.cilium.io/
helm repo update

# 安装 Cilium，启用 kube-proxy replacement 与 FQDN 支持
helm install cilium cilium/cilium --version 1.14.5 \
  --namespace kube-system \
  --set kubeProxyReplacement=strict \
  --set k8sServiceHost=API_SERVER_IP \
  --set k8sServicePort=6443 \
  --set dnsProxy.enable=true \
  --set extraConfig.enable-envoy-config=true
```

#### 3.4.2 部署 argus-gateway

详见 [helm-install.md](./helm-install.md)。安装完成后确认 `argus-gateway` Service 已就绪：

```bash
kubectl -n argus-system get svc argus-gateway
kubectl -n argus-system get pods -l app.kubernetes.io/component=gateway
```

#### 3.4.3 创建 CiliumEgressGatewayPolicy

通过 Helm 安装 Argus 时启用 Cilium 引流方案：

```bash
helm install argus ./deploy/helm/argus \
  --namespace argus-system \
  --create-namespace \
  --set trafficInterception.method=cilium \
  --set trafficInterception.cilium.gatewayService=argus-gateway
```

Helm 渲染出的 `CiliumEgressGatewayPolicy` 示例（来自 `templates/cilium/egress-policy.yaml`）：

```yaml
apiVersion: cilium.io/v2
kind: CiliumEgressGatewayPolicy
metadata:
  name: argus-llm-egress
  namespace: argus-system
spec:
  selectors:
    - podSelector:
        matchLabels: {}
      namespaceSelector:
        matchExpressions:
          - key: kubernetes.io/metadata.name
            operator: NotIn
            values:
              - kube-system
              - argus-system
  destinationCIDRs:
    - 0.0.0.0/0
  destinationPorts:
    - port: "443"
      protocol: TCP
  egressGateway:
    egressIP: argus-gateway.argus-system.svc.cluster.local
```

#### 3.4.4 配置 LLMProvider FQDN 白名单

Argus 通过 `LLMProvider` CRD 声明需要拦截的域名，Cilium DNS proxy 会基于此白名单匹配 FQDN：

```yaml
apiVersion: argus.argus.io/v1alpha1
kind: LLMProvider
metadata:
  name: openai
  namespace: argus-system
spec:
  type: openai-compatible
  hosts:
    - api.openai.com
    - "*.openai.com"
  sni:
    - api.openai.com
  upstream:
    scheme: https
  adapter: openai
  models:
    - gpt-4o
    - gpt-4o-mini
```

### 3.5 验证步骤

```bash
# 1. 确认 Cilium 状态健康
cilium status --wait

# 2. 确认 CiliumEgressGatewayPolicy 已生效
kubectl get cegp -n argus-system argus-llm-egress -o yaml

# 3. 部署测试业务 Pod（不修改任何配置）
kubectl run test-pod --image=curlimages/curl:8.5.0 --rm -it --restart=Never -- \
  curl -v https://api.openai.com/v1/models

# 4. 查看 argus-gateway 日志，确认收到流量
kubectl -n argus-system logs -l app.kubernetes.io/component=gateway -f

# 5. 确认 AIEvent 已落盘
kubectl -n argus-system exec deploy/argus-controller -- \
  ls /var/lib/argus/events/$(date +%Y-%m-%d)/
```

预期结果：

- 业务 Pod `curl` 请求成功返回（200 或 401 等上游响应）。
- `argus-gateway` 日志显示收到来自业务 Pod IP 的连接。
- `argus-controller` 持久化目录出现 `AIEvent` 文件。

### 3.6 局限性

| 局限项 | 说明 | 缓解措施 |
| --- | --- | --- |
| CNI 互斥 | Cilium 与 Calico / 云厂商 ENI CNI 互斥 | 部署前确认 CNI 现状，必要时切换 CNI |
| 内核版本要求 | 内核 ≥ 5.4（建议 ≥ 5.10），老内核不支持 | 升级节点内核或选择 iptables 兜底方案 |
| 托管 K8s 限制 | GKE Autopilot、EKS Fargate 等不支持自定义 eBPF | 切换到标准 GKE / EKS 或自建集群 |
| 动态 IP 依赖 FQDN | LLM 厂商 IP 动态变化时需依赖 DNS proxy | 确保 Cilium DNS proxy 长期运行，配置合理的 TTL |
| Cilium 升级风险 | Cilium 大版本升级可能影响 eBPF 程序兼容性 | 升级前在测试集群验证，主集群灰度升级 |

### 3.7 回滚步骤

```bash
# 1. 卸载 Argus Helm Chart
helm uninstall argus -n argus-system

# 2. 删除 CiliumEgressGatewayPolicy（如未被 Helm 自动清理）
kubectl delete cegp -n argus-system argus-llm-egress --ignore-not-found

# 3. 删除 Argus 相关 CRD
kubectl delete crd argussecuritypolicies.argus.argus.io llmproviders.argus.argus.io --ignore-not-found

# 4. 删除 Argus 命名空间
kubectl delete namespace argus-system --ignore-not-found

# 5. 验证业务 Pod 恢复直连
kubectl run test-pod --image=curlimages/curl:8.5.0 --rm -it --restart=Never -- \
  curl -v https://api.openai.com/v1/models

# 6. （可选）如需完全卸载 Cilium，参考 Cilium 官方卸载文档
# helm uninstall cilium -n kube-system
# 注意：卸载 Cilium 会断开集群所有 Pod 网络，请谨慎操作
```

## 4. iptables / nftables TPROXY 详细方案（MVP 兜底）

### 4.1 方案原理

在节点上通过 `iptables -t mangle TPROXY` 将目的为 LLM IP/端口的流量透明代理到本机 `argus-gateway` 监听端口。`TPROXY` 通过 `IP_TRANSPARENT` socket 选项让网关以原始目的 IP 接收连接，从而保留业务 Pod 视角的"直连"语义。

### 4.2 优势

- 通用、无 CNI 绑定、内核原生支持。
- 适用于任何 Linux 节点，无内核版本强约束（建议 ≥ 4.19）。
- 与 kube-proxy iptables 模式可共存（需小心链顺序）。

### 4.3 前置条件

| 前置项 | 要求 | 验证命令 |
| --- | --- | --- |
| 内核版本 | ≥ 4.19（建议 ≥ 5.4） | `uname -r` |
| 内核模块 | `xt_TPROXY`、`nf_conntrack` 已加载 | `lsmod \| grep -E 'TPROXY\|conntrack'` |
| 节点权限 | 节点 root 或 `CAP_NET_ADMIN` | DaemonSet 配置 `securityContext.privileged: true` |
| DNS 解析前置 | 需配套 IP 列表同步机制（轮询 LLMProvider.hosts） | Argus controller 定期解析 |
| 多副本网关 | 需在节点间做 DNAT 分流（基于 consistent hash 或 BGP/keepalived VIP） | DaemonSet 中配置 |
| kube-proxy 模式 | 与 iptables 模式共存需小心链顺序；建议 kube-proxy 切换到 ipvs 模式 | `kubectl -n kube-system get cm kube-proxy -o yaml \| grep mode` |

### 4.4 安装步骤

#### 4.4.1 部署 argus-gateway

详见 [helm-install.md](./helm-install.md)。网关需监听 TPROXY 端口（默认 8443），并启用 `IP_TRANSPARENT` socket 选项。

#### 4.4.2 安装 Argus 时启用 iptables 引流方案

```bash
helm install argus ./deploy/helm/argus \
  --namespace argus-system \
  --create-namespace \
  --set trafficInterception.method=iptables \
  --set trafficInterception.iptables.tproxyPort=8443 \
  --set trafficInterception.iptables.llmCIDRs="{0.0.0.0/0}" \
  --set trafficInterception.iptables.llmPorts="{443}"
```

Helm 渲染出的 `DaemonSet/argus-node-tproxy`（来自 `templates/iptables/ds-node-tproxy.yaml`）核心字段：

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: argus-node-tproxy
  namespace: argus-system
spec:
  selector:
    matchLabels:
      app.kubernetes.io/component: node-tproxy
  template:
    metadata:
      labels:
        app.kubernetes.io/component: node-tproxy
    spec:
      hostNetwork: true
      hostPID: false
      serviceAccountName: argus-tproxy
      containers:
        - name: tproxy
          image: argus/tproxy-agent:0.1.0
          securityContext:
            privileged: true
            capabilities:
              add: ["NET_ADMIN", "NET_RAW"]
          env:
            - name: TPROXY_PORT
              value: "8443"
            - name: GATEWAY_SERVICE
              value: "argus-gateway.argus-system.svc.cluster.local"
          volumeMounts:
            - name: iptables-init
              mountPath: /scripts/iptables-init.sh
              subPath: iptables-init.sh
      volumes:
        - name: iptables-init
          configMap:
            name: argus-tproxy-rules
            defaultMode: 0755
```

> 注：`DaemonSet/argus-node-tproxy` 是**节点级引流组件**，运行在 `argus-system` 命名空间，仅做透明流量重定向，**不是业务 Pod 的 sidecar / sensor**，符合 C2。

#### 4.4.3 节点 iptables 规则初始化

`ConfigMap/argus-tproxy-rules` 中的 `iptables-init.sh` 示例：

```bash
#!/bin/bash
set -euo pipefail

TPROXY_PORT="${TPROXY_PORT:-8443}"
MARK=0x1

# 创建 mangle 链
iptables -t mangle -N ARGUS_DIVERT 2>/dev/null || iptables -t mangle -F ARGUS_DIVERT
iptables -t mangle -N ARGUS_TPROXY 2>/dev/null || iptables -t mangle -F ARGUS_TPROXY

# 标记本地产生的、目的为 LLM 端口的流量
iptables -t mangle -A PREROUTING -p tcp --dport 443 -j ARGUS_TPROXY

# 排除 argus-system 自身
iptables -t mangle -A ARGUS_TPROXY -m mark --mark 0x1 -j RETURN
iptables -t mangle -A ARGUS_TPROXY -m owner --uid-owner 1000 -j RETURN  # argus-gateway uid

# TPROXY 重定向
iptables -t mangle -A ARGUS_TPROXY -p tcp -j TPROXY \
  --tproxy-mark ${MARK} \
  --on-port ${TPROXY_PORT} \
  --on-ip 0.0.0.0

# 路由表：标记流量走本地 lo
ip rule add fwmark ${MARK} table 100 2>/dev/null || true
ip route add local 0.0.0.0/0 dev lo table 100 2>/dev/null || true

echo "[argus-tproxy] iptables rules installed"
```

#### 4.4.4 配套 IP 列表同步

`argus-controller` 定期（默认 60s）解析 `LLMProvider.hosts` 得到 IP 集合，通过 `iptables -t mangle -R` 或 `nftables set` 更新匹配规则。多副本网关时通过 consistent hash 将源 IP 分配到不同 gateway Pod。

### 4.5 验证步骤

```bash
# 1. 确认 DaemonSet 在每个节点运行
kubectl -n argus-system get ds argus-node-tproxy
kubectl -n argus-system get pods -l app.kubernetes.io/component=node-tproxy -o wide

# 2. 在节点上验证 iptables 规则
kubectl -n argus-system exec ds/argus-node-tproxy -- iptables -t mangle -L ARGUS_TPROXY -n -v

# 3. 部署测试业务 Pod
kubectl run test-pod --image=curlimages/curl:8.5.0 --rm -it --restart=Never -- \
  curl -v https://api.openai.com/v1/models

# 4. 查看 argus-gateway 日志
kubectl -n argus-system logs -l app.kubernetes.io/component=gateway -f

# 5. 确认 AIEvent 已落盘
kubectl -n argus-system exec deploy/argus-controller -- \
  ls /var/lib/argus/events/$(date +%Y-%m-%d)/
```

### 4.6 局限性

| 局限项 | 说明 | 缓解措施 |
| --- | --- | --- |
| 无法基于 SNI 精准匹配 | iptables 工作在 L3/L4，看不到 TLS SNI | 前置 DNS 解析：controller 定期解析 LLMProvider.hosts 得到 IP 集合 |
| LLM 厂商 IP 频繁变化 | DNS TTL 过期后 IP 可能变化 | controller 60s 轮询 DNS 更新规则集 |
| 多副本网关分流复杂 | 多副本时需在节点间做 DNAT 分流 | 基于 consistent hash 或 BGP/keepalived VIP |
| 与 kube-proxy iptables 模式共存 | 链顺序冲突可能导致规则失效 | 切换 kube-proxy 到 ipvs 模式，或显式调整链优先级 |
| 节点资源占用 | 每节点一个 DaemonSet，占用少量 CPU/内存 | 资源 requests/limits 严格声明 |
| 跨节点流量 | 业务 Pod 与 gateway Pod 不在同一节点时需跨节点转发 | 配合 `nodeSelector` 或 `podAntiAffinity` 优化调度 |

### 4.7 回滚步骤

```bash
# 1. 卸载 Argus Helm Chart（自动清理 DaemonSet 与 ConfigMap）
helm uninstall argus -n argus-system

# 2. 在每个节点清理残留 iptables 规则（DaemonSet 卸载时未自动清理）
for node in $(kubectl get nodes -o name); do
  kubectl debug node/${node#node/} --image=alpine -- chroot /host \
    iptables -t mangle -F ARGUS_TPROXY 2>/dev/null || true
  kubectl debug node/${node#node/} --image=alpine -- chroot /host \
    iptables -t mangle -F ARGUS_DIVERT 2>/dev/null || true
  kubectl debug node/${node#node/} --image=alpine -- chroot /host \
    iptables -t mangle -X ARGUS_TPROXY 2>/dev/null || true
  kubectl debug node/${node#node/} --image=alpine -- chroot /host \
    iptables -t mangle -X ARGUS_DIVERT 2>/dev/null || true
  kubectl debug node/${node#node/} --image=alpine -- chroot /host \
    ip rule del fwmark 0x1 table 100 2>/dev/null || true
  kubectl debug node/${node#node/} --image=alpine -- chroot /host \
    ip route del local 0.0.0.0/0 dev lo table 100 2>/dev/null || true
done

# 3. 删除 Argus CRD 与命名空间
kubectl delete crd argussecuritypolicies.argus.argus.io llmproviders.argus.argus.io --ignore-not-found
kubectl delete namespace argus-system --ignore-not-found

# 4. 验证业务 Pod 恢复直连
kubectl run test-pod --image=curlimages/curl:8.5.0 --rm -it --restart=Never -- \
  curl -v https://api.openai.com/v1/models
```

## 5. Istio 经典 sidecar 不可选的原因

### 5.1 经典 Istio 数据平面架构

经典 Istio（1.x 默认模式）通过 `istioctl install` 或 `IstioOperator` 在每个业务 Pod 中注入 Envoy sidecar，由 sidecar 拦截 Pod 的所有入站与出站流量，再根据 `VirtualService` / `ServiceEntry` 路由到 Egress Gateway 或直接放行。

### 5.2 与 C2 的直接冲突

spec §2 中 C2 明确规定：

> **禁止 Sidecar / DaemonSet**：禁止 Pod 侧车模式、禁止每容器 Agent、禁止节点级 DaemonSet 传感器。

经典 Istio 的 sidecar 注入属于典型的"Pod 侧车模式"，每业务 Pod 都会被注入一个 Envoy 容器，直接违反 C2。即便使用 `istioctl` 的 `--set values.global.proxy.privileged=true` 或自定义模板，也无法改变"业务 Pod 内多出一个 sidecar 容器"这一事实。

### 5.3 Ambient 模式的可行性评估

Istio 1.18+ 引入 Ambient Mesh 模式（ztunnel + waypoint），可在零 sidecar 的前提下实现 L4 流量拦截。但 MVP 阶段不作为首选，原因如下：

| 评估维度 | 状态 | 说明 |
| --- | --- | --- |
| 生产稳定性 | 待评估 | Ambient 模式仍处于较新阶段，生产案例较少 |
| L7 流量识别能力 | 受限 | ztunnel 仅做 L4，L7 需要 waypoint 代理，架构复杂度上升 |
| 与 C2 兼容性 | 兼容 | ztunnel 以 DaemonSet 形式运行，但**仅做 L4 透明转发，不做检测**，与 C2 中"禁止节点级 DaemonSet 传感器"的"传感器"语义不冲突 |
| 集成成本 | 高 | 需要额外维护 Istio 控制面、waypoint 配置、与 Cilium 共存策略 |
| MVP 适配性 | 不推荐 | 已有 Cilium 与 iptables 两方案覆盖 MVP 需求，引入 Istio 增加复杂度 |

### 5.4 结论

经典 Istio Sidecar 方案**直接违背 C2，不可选**。Ambient 模式可作为后续迭代的备选方案，MVP 阶段不引入。

## 6. 云厂商 VPC 路由方案说明（仅备选）

### 6.1 方案原理

在 VPC 路由表将 LLM 厂商 CIDR 指向 `argus-gateway` 所在 ENI。流量在 VPC 网络层被路由到网关节点，由网关节点上的 iptables/路由规则再转发到 `argus-gateway` Pod。

### 6.2 硬性限制

- 仅按 CIDR 路由，无法基于域名。
- LLM 厂商 IP 不固定且无官方 CIDR 列表，需自行维护。
- 跨可用区流量产生费用。
- 路由变更影响整张 VPC，风险高。

### 6.3 定位

仅作为兜底兜底，不建议 MVP 使用。在 Cilium 与 iptables 均不可用的极端场景下考虑。

## 7. MVP 选型结论

| 项 | 决定 |
| --- | --- |
| 首选方案 | Cilium eBPF（FQDN + Pod identity 重定向） |
| 兜底方案 | iptables/nftables TPROXY |
| 不可选 | 经典 Istio Sidecar（违背 C2） |
| 后续评估 | Istio Ambient Mesh（待生产成熟度评估） |
| 极端兜底 | 云厂商 VPC 路由（不建议 MVP 使用） |
| 文档要求 | 部署文档需写明两种 MVP 方案的安装步骤、前置条件、局限性、回滚步骤 |

### 7.1 选型决策树

```
集群是否使用 Cilium 作为 CNI？
├── 是
│   └── 内核版本 ≥ 5.4？
│       ├── 是 → 选择 Cilium eBPF（首选）
│       └── 否 → 选择 iptables TPROXY（兜底）
└── 否
    └── 是否可切换到 Cilium？
        ├── 是 → 切换 CNI 后选择 Cilium eBPF
        └── 否 → 选择 iptables TPROXY（兜底）
```

### 7.2 Helm values 配置示例

详见 [helm-install.md](./helm-install.md) 第 4 节"引流方案选择"。简要示例：

```yaml
# Cilium 首选
trafficInterception:
  method: cilium
  cilium:
    gatewayService: argus-gateway

# iptables 兜底
trafficInterception:
  method: iptables
  iptables:
    tproxyPort: 8443
    llmCIDRs:
      - 0.0.0.0/0
    llmPorts:
      - 443
```

## 8. 引流层故障的降级行为

引流层本身故障（如 Cilium eBPF 程序异常、iptables 规则被误删）属于 spec §12.3 中的"引流层故障"场景：

| 场景 | monitor + fail-open | monitor + fail-closed | enforce + fail-open | enforce + fail-closed |
| --- | --- | --- | --- | --- |
| 引流层故障（流量未到达网关） | 业务直连 LLM（兜底） | 业务直连 LLM（兜底） | 业务直连 LLM（兜底） | 业务直连 LLM（兜底） |

引流层故障是**安全降级**而非"故障关"：业务连通性优先，但需通过 Prometheus 告警 `argus_gateway_health=0` 与 `cilium_policy_count` 等指标触发告警。详见 [runmodes-failure.md](./runmodes-failure.md)。

## 9. 相关文档

- [架构设计](./architecture.md)
- [Pod 身份溯源设计](./pod-identity.md)
- [TLS 解密检测设计](./tls-design.md)
- [运行模式与故障场景](./runmodes-failure.md)
- [Helm 安装指南](./helm-install.md)
- [端到端验证](./e2e-verification.md)
