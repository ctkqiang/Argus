# Argus 观枢 Makefile
# 维护者：Argus 核心开发团队
# 说明：所有目标均可在项目根目录通过 make <target> 调用。
# 详细规范见 .trae/rules/go-coding-standards.md 附录 A。

# === 基础变量声明 ===
BINARY_GATEWAY    := argus-gateway
BINARY_CONTROLLER := argus-controller
MODULE            := github.com/ctkqiang/argus
OUTDIR            := bin

# === 开发环境变量 ===
CONFIG_GATEWAY    ?= configs/gateway.yaml.example
CONFIG_CONTROLLER ?= configs/controller.yaml.example
RUN_ARGS          ?=
DEV_CLUSTER       ?= argus-dev
DEV_NS            ?= argus
GATEWAY_PORT      ?= 8443
GATEWAY_METRICS   ?= 19090
CONTROLLER_GRPC   ?= 18444
CONTROLLER_METRICS?= 19091

# === 工具链变量 ===
GO             ?= go

# === 默认目标 ===
.PHONY: all
all: fmt lint test build

.PHONY: help
help:
	@echo "Argus (观枢) Makefile 可用目标："
	@echo ""
	@echo "── 开发环境（一键启动）──"
	@echo "  run               - 一键启动 Minikube + 构建镜像 + 部署 + 端口转发（默认入口）"
	@echo "  dev-up            - 同 make run"
	@echo "  kill              - 停止开发环境（保留集群数据，下次 make run 秒恢复）"
	@echo "  dev-down          - 停止并删除 Minikube 开发集群（彻底释放磁盘）"
	@echo "  dev-status        - 查看开发环境 Pod/Service 状态"
	@echo "  dev-logs          - 实时查看所有 Pod 日志（tail）"
	@echo "  dev-rebuild       - 重新构建镜像并滚动更新（代码改了跑这个）"
	@echo ""
	@echo "── 本地直接运行（不依赖 K8s）──"
	@echo "  run-gateway       - go run 启动 gateway（CONFIG=路径 可覆盖配置）"
	@echo "  run-controller    - go run 启动 controller（CONFIG=路径 可覆盖配置）"
	@echo ""
	@echo "── 构建与测试──"
	@echo "  all               - 顺序执行 fmt → lint → test → build（默认目标）"
	@echo "  fmt               - 格式化 Go 源码（gofmt + 若存在 goimports）"
	@echo "  lint              - 静态检查（go vet ./...，若存在则追加 staticcheck）"
	@echo "  test              - 运行全部单元测试，开启竞态检测与覆盖率统计"
	@echo "  build             - 编译 gateway 和 controller 两个二进制（输出到 $(OUTDIR)/）"
	@echo "  build-gateway     - 仅编译 $(BINARY_GATEWAY)"
	@echo "  build-controller  - 仅编译 $(BINARY_CONTROLLER)"
	@echo "  clean             - 清理构建产物（$(OUTDIR)/）"
	@echo "  tidy              - 整理 go.mod 与 go.sum 依赖"
	@echo ""
	@echo "── 代码生成──"
	@echo "  proto             - 基于 buf 生成 Proto 桩代码（调用 scripts/gen-proto.sh）"
	@echo "  crd               - 基于 controller-gen 生成 DeepCopy 与 CRD YAML"
	@echo ""

# ── 一键开发环境 ──

.PHONY: run dev-up
run: dev-up
dev-up:
	@bash scripts/dev-up.sh

.PHONY: kill
kill:
	@bash scripts/dev-kill.sh

.PHONY: dev-down
dev-down:
	@bash scripts/dev-down.sh

.PHONY: dev-status
dev-status:
	@echo "[dev-status] Minikube 集群: $(DEV_CLUSTER)"
	@minikube status -p $(DEV_CLUSTER) 2>/dev/null || echo "  集群未运行"
	@echo ""
	@echo "[dev-status] Pod 状态（namespace=$(DEV_NS)）:"
	@kubectl get pods -n $(DEV_NS) -o wide 2>/dev/null || echo "  namespace $(DEV_NS) 不存在或无法连接"
	@echo ""
	@echo "[dev-status] Service 状态:"
	@kubectl get svc -n $(DEV_NS) 2>/dev/null || true

.PHONY: dev-logs
dev-logs:
	@kubectl logs -f -n $(DEV_NS) -l app.kubernetes.io/name=argus --all-containers=true --prefix=true --tail=50

.PHONY: dev-rebuild
dev-rebuild:
	@echo "[dev-rebuild] 重新构建镜像并滚动更新..."
	@eval $$(minikube docker-env -p $(DEV_CLUSTER)) && \
		docker build --target gateway -t argus-gateway:dev --build-arg "VERSION=dev-$$(date +%s)" -q . && \
		docker build --target controller -t argus-controller:dev --build-arg "VERSION=dev-$$(date +%s)" -q . && \
		echo "[dev-rebuild] 镜像构建完成，触发滚动重启..." && \
		kubectl rollout restart deployment/argus-gateway -n $(DEV_NS) && \
		kubectl rollout restart deployment/argus-controller -n $(DEV_NS) && \
		kubectl rollout status deployment/argus-gateway -n $(DEV_NS) --timeout=120s && \
		kubectl rollout status deployment/argus-controller -n $(DEV_NS) --timeout=120s && \
		echo "[dev-rebuild] 滚动更新完成"

# ── 本地直接运行 ──

.PHONY: run-gateway
run-gateway:
	@echo "[run-gateway] 启动 $(BINARY_GATEWAY)，配置: $${CONFIG:-$(CONFIG_GATEWAY)}"
	@$(GO) run ./cmd/argus-gateway -config $${CONFIG:-$(CONFIG_GATEWAY)} $(RUN_ARGS)

.PHONY: run-controller
run-controller:
	@echo "[run-controller] 启动 $(BINARY_CONTROLLER)，配置: $${CONFIG:-$(CONFIG_CONTROLLER)}"
	@$(GO) run ./cmd/argus-controller -config $${CONFIG:-$(CONFIG_CONTROLLER)} $(RUN_ARGS)

# ── 构建 ──

.PHONY: fmt
fmt:
	@echo "[fmt] 执行 gofmt -s -w ."
	@gofmt -s -w .
	@if command -v goimports >/dev/null 2>&1; then \
		echo "[fmt] 执行 goimports -w ."; \
		goimports -w .; \
	else \
		echo "[fmt] 提示：goimports 未安装，建议执行：go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

.PHONY: lint
lint:
	@echo "[lint] 执行 go vet ./..."
	@$(GO) vet ./...
	@if command -v staticcheck >/dev/null 2>&1; then \
		echo "[lint] 执行 staticcheck ./..."; \
		staticcheck ./...; \
	else \
		echo "[lint] 提示：staticcheck 未安装，建议执行：go install honnef.co/go/tools/cmd/staticcheck@latest"; \
	fi

.PHONY: test
test:
	@echo "[test] 执行 go test -race -cover ./..."
	@$(GO) test -race -cover ./...

.PHONY: build
build: build-gateway build-controller
	@echo "[build] 完成，二进制输出到 $(OUTDIR)/ 目录"

.PHONY: build-gateway
build-gateway:
	@echo "[build-gateway] 编译 $(BINARY_GATEWAY)"
	@mkdir -p $(OUTDIR)
	@$(GO) build -o $(OUTDIR)/$(BINARY_GATEWAY) ./cmd/argus-gateway

.PHONY: build-controller
build-controller:
	@echo "[build-controller] 编译 $(BINARY_CONTROLLER)"
	@mkdir -p $(OUTDIR)
	@$(GO) build -o $(OUTDIR)/$(BINARY_CONTROLLER) ./cmd/argus-controller

.PHONY: proto
proto:
	@if command -v buf >/dev/null 2>&1; then \
		echo "[proto] 执行 scripts/gen-proto.sh"; \
		bash scripts/gen-proto.sh; \
	else \
		echo "[proto] 错误：buf 未安装，请执行：go install github.com/bufbuild/buf/cmd/buf@latest"; \
		exit 1; \
	fi

.PHONY: crd
crd:
	@if command -v controller-gen >/dev/null 2>&1; then \
		echo "[crd] 执行 scripts/gen-crd.sh"; \
		echo "[crd] 步骤 1/2: controller-gen object paths=\"./...\" output:dir=."; \
		controller-gen object paths="./..." output:dir=.; \
		echo "[crd] 步骤 2/2: 生成 CRD YAML 到 deploy/helm/argus/templates/crds"; \
		bash scripts/gen-crd.sh; \
	else \
		echo "[crd] 错误：controller-gen 未安装，请执行：go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest"; \
		exit 1; \
	fi

.PHONY: clean
clean:
	@echo "[clean] 清理 $(OUTDIR)/"
	@rm -rf $(OUTDIR)/

.PHONY: tidy
tidy:
	@echo "[tidy] 执行 go mod tidy"
	@$(GO) mod tidy
