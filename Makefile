# Argus 观枢 Makefile
# 维护者：ctkqiang
# 说明：所有目标均可在项目根目录通过 make <target> 调用。

# === 工具链变量 ===
GO             ?= go
BUF            ?= buf
CONTROLLER_GEN ?= controller-gen
GOLANGCI_LINT  ?= golangci-lint

# === 路径变量 ===
BIN_DIR        ?= bin
PKGS           := ./...
PROTO_DIR      := proto

# === 默认目标 ===
.PHONY: all
all: build

# build: 编译 argus-gateway 与 argus-controller 二进制到 bin/。
.PHONY: build
build:
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/argus-gateway ./cmd/argus-gateway
	$(GO) build -o $(BIN_DIR)/argus-controller ./cmd/argus-controller

# test: 运行全部单元测试，开启竞态检测与覆盖率统计。
.PHONY: test
test:
	$(GO) test -race -cover $(PKGS)

# proto: 基于 buf 从 proto/ 生成 Go 桩代码到 api/proto/argus/。
.PHONY: proto
proto:
	cd $(PROTO_DIR) && $(BUF) generate

# crd: 基于 controller-gen 生成 DeepCopy 方法与 CRD YAML。
.PHONY: crd
crd: object
	$(CONTROLLER_GEN) crd paths=./api/argus/v1alpha1/... output:crd:artifacts:config=deploy/helm/argus/templates/crds

# object: 生成 zz_generated.deepcopy.go。
.PHONY: object
object:
	$(CONTROLLER_GEN) object paths=./api/argus/v1alpha1/...

# lint: 静态检查，优先 golangci-lint，未安装时回退到 go vet。
.PHONY: lint
lint:
	$(GO) vet $(PKGS)
	@command -v $(GOLANGCI_LINT) >/dev/null 2>&1 && $(GOLANGCI_LINT) run $(PKGS) || echo "[lint] golangci-lint 未安装，仅执行 go vet"

# e2e: 运行端到端测试（待后续任务实现）。
.PHONY: e2e
e2e:
	@echo "[e2e] 端到端测试待实现，参见 test/e2e/"

# tidy: 整理 go.mod 与 go.sum。
.PHONY: tidy
tidy:
	$(GO) mod tidy

# install-tools: 安装开发工具链到 GOBIN。
.PHONY: install-tools
install-tools:
	$(GO) install github.com/bufbuild/buf/cmd/buf@latest
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	$(GO) install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# clean: 清理构建产物。
.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
