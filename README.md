# Argus（观枢）

Kubernetes 原生 LLM 出站流量安全网关。

业务 Pod 不改一行代码，集群级透明引流，拦截所有发往大模型厂商的出口请求。对 Prompt 做提取、多检测器流水线、风险打分、放行或阻断，事件归档成 AIEvent。

当前进度：**阶段一 1-A**（项目骨架）。gRPC 契约、CRD 类型、构建链路、K8s 部署、一键开发环境都在。业务逻辑（1-B/C/D）还没写，子包里先放了 `doc.go` 占位。

---

## 做什么

- 集群级透明拦截所有向 LLM 厂商发的 HTTPS 请求（Cilium eBPF / iptables TPROXY）
- 非 LLM 流量直接透传，零额外延迟
- Prompt 过四段检测流水线：rules → heuristic → encoding → semantic，输出 risk_score
- 按 `ArgusSecurityPolicy` CRD 决策 allow / block / degraded，支持 monitor 和 enforce 模式
- 事件流式上报 controller，落盘 JSONL（at-least-once）
- Pod 身份解析走 controller，gateway 不碰 k8s 权限

## 不做（阶段一）

- 多厂商 adapter 真实逻辑（只有 OpenAI 兼容骨架）
- 语义检测器接外部 LLM（接口预留，返回 unimplemented）
- 可视化面板、C++ 检测器、多集群联邦、对象存储 / SIEM 直写

---

## 目录结构

```
argus/
├── api/argus/v1alpha1/           CRD Go 类型 + DeepCopy
├── cmd/
│   ├── argus-gateway/            数据平面入口（/healthz /readyz + 信号退出）
│   └── argus-controller/         控制平面入口
├── configs/                      配置样例
├── deploy/
│   ├── examples/                 LLMProvider / ArgusSecurityPolicy 样例
│   └── helm/argus/               Helm Chart
├── docs/                         设计文档 + 文档站（React + Ant Design）
├── internal/
│   ├── config/                   配置加载（viper + env + 默认值）
│   ├── detector/{rules,heuristic,encoding,semantic}/
│   ├── gateway/{server,tls,proxy,protocol,prompt,adapter,pipeline,risk,policy,identity,metrics,event}/
│   ├── controller/{policy,identity,health,event,metrics}/
│   └── utilities/                slog 日志（纯标准库）
├── k8s/                          Kustomize base + dev/stg/prod overlays
├── pkg/
│   ├── pb/argus/*/v1alpha1/      protobuf + gRPC 生成桩
│   └── {apierrors,signal,logger,tracing}/
├── proto/argus/                  5 份 gRPC 契约源文件
├── scripts/                      构建 / 开发环境脚本
├── terraform/                    AWS EKS / GCP GKE / Azure AKS 模块
├── Dockerfile                    多阶段构建 → distroless
├── Makefile                      所有构建目标
└── README.md
```

---

## 环境要求

- Go >= 1.22
- Kubernetes >= 1.27
- Helm >= 3.10
- Docker Desktop（开发环境）
- kubectl + minikube（开发环境）

```bash
# 工具链
go install golang.org/x/tools/cmd/goimports@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
go install github.com/bufbuild/buf/cmd/buf@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
```

---

## 快速开始

### 一键开发环境

装好 Docker、kubectl、minikube，然后：

```bash
make run
```

脚本会自动检测 Docker 可用资源（CPU / 内存），按安全余量分配给 Minikube，不会因为内存超限炸掉。启动完成后打印访问地址：

```
Gateway Proxy:      https://localhost:8443
Gateway Metrics:    http://localhost:19090/metrics
Gateway Health:     http://localhost:19090/healthz
Controller gRPC:    localhost:18444
Controller Metrics: http://localhost:19091/metrics
```

日常操作：

```bash
make dev-status     # Pod / Service 状态
make dev-logs       # 实时日志
make dev-rebuild    # 改了代码，重新构建 + 滚动更新
make kill           # 停止环境，保留集群数据（下次 make run 秒恢复）
make dev-down       # 彻底删除集群
```

手动覆盖资源：

```bash
ARGUS_CPUS=2 ARGUS_MEMORY=1536 make run
```

### 本地直接跑（不依赖 K8s）

```bash
make run-gateway
make run-controller
```

---

## 构建

```bash
make build    # 产物：bin/argus-gateway + bin/argus-controller
```

Docker 镜像（多阶段构建，distroless 运行时，无 shell 无包管理器）：

```bash
docker build --target gateway    -t argus-gateway:dev    .
docker build --target controller -t argus-controller:dev .
```

| 镜像 | 基础镜像 | 端口 | 用户 |
|---|---|---|---|
| argus-gateway | distroless/static:nonroot | 8443 / 9090 | nonroot |
| argus-controller | distroless/static:nonroot | 8444 / 9090 | nonroot |

---

## K8s 部署

```bash
kubectl apply -k k8s/overlays/dev     # 开发
kubectl apply -k k8s/overlays/prod    # 生产（多副本）
```

| overlay | 说明 |
|---|---|
| dev | 单副本，本地镜像，imagePullPolicy: Never |
| stg | 资源限制（1C/1Gi ~ 4C/4Gi） |
| prod | gateway x3, controller x2 |

---

## 代码生成

```bash
make proto    # buf generate → pkg/pb/argus/*/v1alpha1/*.pb.go
make crd      # controller-gen → CRD YAML + DeepCopy
```

---

## 提交前

```bash
make all      # fmt → lint → test → build
```

或者分步：

```bash
make fmt      # gofmt + goimports
make lint     # go vet + staticcheck
make test     # go test -race -cover（普通包 ≥ 70%，detector ≥ 90%）
```

---

## 健康检查

gateway 和 controller 都暴露：

| 路径 | 用途 |
|---|---|
| `/healthz` | Liveness Probe，返回 200 |
| `/readyz` | Readiness Probe，返回 200 |

---

## 样例

```bash
kubectl create secret generic openai-api-key --from-literal=api_key='sk-xxxx'
kubectl apply -f deploy/examples/llmprovider-openai.yaml
kubectl apply -f deploy/examples/argussecuritypolicy-default.yaml
```

---

## 故障排查

**IDE 报 `Import "argus/detection.proto" was not found`，但 `make proto` 是绿的**

IDE 的 proto import root 没指到 `proto/` 目录。跑：

```bash
bash scripts/fix-proto-imports.sh
```

跑完还红 → IDE 里手动把 proto include root 设成 `<project>/proto`。

**`go build` 报 `no required module provides package .../pkg/pb/...`**

```bash
make proto && go mod tidy && go build ./...
```

**`make run` 报 Minikube 内存不够**

脚本已经会自动适配。如果还是不行，手动指定：

```bash
ARGUS_MEMORY=1536 make run
```

或者直接：

```bash
minikube start -p argus-dev --cpus=2 --memory=1536
```

---

## Makefile 目标一览

| 目标 | 说明 |
|---|---|
| `make run` | 一键 Minikube 开发环境 |
| `make kill` | 停止环境，保留数据 |
| `make dev-down` | 删除集群 |
| `make dev-status` | Pod / Service 状态 |
| `make dev-logs` | 实时日志 |
| `make dev-rebuild` | 重新构建 + 滚动更新 |
| `make run-gateway` | 本地 go run gateway |
| `make run-controller` | 本地 go run controller |
| `make all` | fmt → lint → test → build |
| `make fmt` | 格式化 |
| `make lint` | 静态检查 |
| `make test` | 单元测试 |
| `make build` | 编译二进制 |
| `make proto` | 生成 protobuf 桩 |
| `make crd` | 生成 CRD YAML |
| `make tidy` | go mod tidy |
| `make clean` | 清理 bin/ |

---

## 设计文档

| 文档 | 内容 |
|---|---|
| [architecture.md](docs/architecture.md) | 逻辑拓扑 + 报文流转 13 步 |
| [traffic-interception.md](docs/traffic-interception.md) | 4 种透明引流方案对比 |
| [pod-identity.md](docs/pod-identity.md) | 身份溯源 + 5 场景准确性矩阵 |
| [tls-design.md](docs/tls-design.md) | TLS 解密 + CA 信任 + 降级表 |
| [runmodes-failure.md](docs/runmodes-failure.md) | monitor/enforce × fail-open/closed 矩阵 |

文档站本地预览：

```bash
python3 -m http.server 8080 -d docs/
```

---

## 内部文档

- 代码规范：`.trae/rules/go-coding-standards.md`
- 产品 Spec：`.trae/specs/argus-guanshu/spec.md`
- 任务拆解：`.trae/specs/argus-guanshu/tasks.md`
- 验收清单：`.trae/specs/argus-guanshu/checklist.md`
