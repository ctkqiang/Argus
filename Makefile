# Argus 观枢 Makefile
# 维护者：Argus 核心开发团队
# 说明：所有目标均可在项目根目录通过 make <target> 调用。
# 详细规范见 .trae/rules/go-coding-standards.md 附录 A。

# === 基础变量声明 ===
BINARY_GATEWAY    := argus-gateway
BINARY_CONTROLLER := argus-controller
MODULE            := github.com/ctkqiang/argus
OUTDIR            := bin

# === 工具链变量 ===
GO             ?= go

# === 默认目标 ===
.PHONY: all
all: fmt lint test build

# help: 列出所有可用目标及说明。
.PHONY: help
help:
	@echo "Argus (观枢) Makefile 可用目标："
	@echo ""
	@echo "  all               - 顺序执行 fmt → lint → test → build（默认目标）"
	@echo "  help              - 显示本帮助信息"
	@echo "  fmt               - 格式化 Go 源码（gofmt + 若存在 goimports）"
	@echo "  lint              - 静态检查（go vet ./...，提示安装 staticcheck）"
	@echo "  test              - 运行全部单元测试，开启竞态检测与覆盖率统计"
	@echo "  build             - 编译 gateway 和 controller 两个二进制（输出到 $(OUTDIR)/）"
	@echo "  build-gateway     - 仅编译 $(BINARY_GATEWAY)"
	@echo "  build-controller  - 仅编译 $(BINARY_CONTROLLER)"
	@echo "  proto             - 基于 buf 生成 Proto 桩代码（调用 scripts/gen-proto.sh）"
	@echo "  crd               - 基于 controller-gen 生成 DeepCopy 与 CRD YAML"
	@echo "  clean             - 清理构建产物（$(OUTDIR)/ 与 pkg/pb/）"
	@echo "  tidy              - 整理 go.mod 与 go.sum 依赖"
	@echo ""

# fmt: 格式化 Go 源码，优先 gofmt，存在 goimports 时追加执行。
.PHONY: fmt
fmt:
	@echo "[fmt] 执行 gofmt -s -w ."
	gofmt -s -w .
	@if command -v goimports >/dev/null 2>&1; then \
		echo "[fmt] 执行 goimports -w ."; \
		goimports -w .; \
	else \
		echo "[fmt] 提示：goimports 未安装，建议执行：go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

# lint: 静态检查，执行 go vet ./...，提示用户可选安装 staticcheck。
.PHONY: lint
lint:
	@echo "[lint] 执行 go vet ./..."
	$(GO) vet ./...
	@if command -v staticcheck >/dev/null 2>&1; then \
		echo "[lint] 执行 staticcheck ./..."; \
		staticcheck ./...; \
	else \
		echo "[lint] 提示：staticcheck 未安装，建议执行：go install honnef.co/go/tools/cmd/staticcheck@latest"; \
	fi

# test: 运行全部单元测试，开启竞态检测与覆盖率统计。
.PHONY: test
test:
	@echo "[test] 执行 go test -race -cover ./..."
	$(GO) test -race -cover ./...

# build: 编译 gateway 和 controller 两个二进制。
.PHONY: build
build: build-gateway build-controller
	@echo "[build] 完成，二进制输出到 $(OUTDIR)/ 目录"

# build-gateway: 仅编译 argus-gateway 二进制。
.PHONY: build-gateway
build-gateway:
	@echo "[build-gateway] 编译 $(BINARY_GATEWAY)"
	@mkdir -p $(OUTDIR)
	$(GO) build -o $(OUTDIR)/$(BINARY_GATEWAY) ./cmd/argus-gateway

# build-controller: 仅编译 argus-controller 二进制。
.PHONY: build-controller
build-controller:
	@echo "[build-controller] 编译 $(BINARY_CONTROLLER)"
	@mkdir -p $(OUTDIR)
	$(GO) build -o $(OUTDIR)/$(BINARY_CONTROLLER) ./cmd/argus-controller

# proto: 调用 scripts/gen-proto.sh 基于 buf 生成 Proto 桩代码。
.PHONY: proto
proto:
	@if command -v buf >/dev/null 2>&1; then \
		echo "[proto] 执行 scripts/gen-proto.sh"; \
		bash scripts/gen-proto.sh; \
	else \
		echo "[proto] 错误：buf 未安装，请执行：go install github.com/bufbuild/buf/cmd/buf@latest"; \
		exit 1; \
	fi

# crd: 调用 scripts/gen-crd.sh，基于 controller-gen 生成 DeepCopy 与 CRD YAML。
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

# clean: 清理构建产物（bin/ 与 pkg/pb/）。
.PHONY: clean
clean:
	@echo "[clean] 清理 $(OUTDIR)/ 与 pkg/pb/"
	rm -rf $(OUTDIR)/
	rm -rf pkg/pb/

# tidy: 整理 go.mod 与 go.sum 依赖。
.PHONY: tidy
tidy:
	@echo "[tidy] 执行 go mod tidy"
	$(GO) mod tidy
