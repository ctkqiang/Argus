#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
STATE_DIR="${PROJECT_ROOT}/.argus-dev"
mkdir -p "${STATE_DIR}"

CLUSTER_NAME="${ARGUS_CLUSTER:-argus-dev}"
NAMESPACE="argus"
GATEWAY_PORT="${ARGUS_GATEWAY_PORT:-8443}"
GATEWAY_METRICS_PORT="${ARGUS_GATEWAY_METRICS_PORT:-19090}"
CONTROLLER_GRPC_PORT="${ARGUS_CONTROLLER_GRPC_PORT:-18444}"
CONTROLLER_METRICS_PORT="${ARGUS_CONTROLLER_METRICS_PORT:-19091}"
WAIT_TIMEOUT="${ARGUS_WAIT_TIMEOUT:-180}"
DISK_SIZE="${ARGUS_DISK_SIZE:-20g}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC}  $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*" >&2; }
step()  { echo -e "\n${CYAN}${BOLD}==>${NC} ${BOLD}$*${NC}"; }

cleanup() {
	echo ""
	info "收到中断信号，正在清理端口转发..."
	if [ -f "${STATE_DIR}/port-forward.pids" ]; then
		while read -r pid; do
			[ -n "$pid" ] && kill "$pid" 2>/dev/null || true
		done < "${STATE_DIR}/port-forward.pids"
		rm -f "${STATE_DIR}/port-forward.pids"
	fi
	info "端口转发已停止。Argus 开发环境仍在 Minikube 中运行。"
	info "如需彻底销毁集群，执行: make dev-down"
	exit 0
}
trap cleanup SIGINT SIGTERM

detect_docker_resources() {
	local mem_bytes cpus
	mem_bytes=$(docker info --format '{{.MemTotal}}' 2>/dev/null || echo "0")
	cpus=$(docker info --format '{{.NCPU}}' 2>/dev/null || echo "1")

	if [ "$mem_bytes" -eq 0 ]; then
		DOCKER_MEM_MB=2048
		DOCKER_CPUS=2
		return
	fi

	DOCKER_MEM_MB=$((mem_bytes / 1024 / 1024))
	DOCKER_CPUS="$cpus"
}

resolve_cluster_resources() {
	local req_cpus req_mem safe_mem

	req_cpus="${ARGUS_CPUS:-}"
	req_mem="${ARGUS_MEMORY:-}"

	if [ -z "$req_cpus" ]; then
		req_cpus=2
		if [ "$DOCKER_CPUS" -le 2 ]; then
			req_cpus="$DOCKER_CPUS"
		elif [ "$DOCKER_CPUS" -le 4 ]; then
			req_cpus=2
		else
			req_cpus=4
		fi
	fi

	if [ -z "$req_mem" ]; then
		safe_mem=$((DOCKER_MEM_MB - 512))
		if [ "$safe_mem" -lt 1536 ]; then
			safe_mem=$((DOCKER_MEM_MB - 256))
		fi
		if [ "$safe_mem" -lt 1024 ]; then
			error "Docker Desktop 可用内存仅 ${DOCKER_MEM_MB}MB，至少需要 1024MB（建议 2GB+）"
			error "请在 Docker Desktop -> Settings -> Resources 中增加内存分配后重试"
			exit 1
		fi
		req_mem="$safe_mem"
		if [ "$req_mem" -gt 2048 ]; then
			req_mem=2048
		fi
	fi

	if [ "$req_cpus" -gt "$DOCKER_CPUS" ]; then
		warn "请求 CPU 数 ($req_cpus) 超过 Docker 可用 ($DOCKER_CPUS)，自动调整为 $DOCKER_CPUS"
		req_cpus="$DOCKER_CPUS"
	fi

	if [ "$req_mem" -ge "$DOCKER_MEM_MB" ]; then
		local adjusted=$((DOCKER_MEM_MB - 256))
		if [ "$adjusted" -lt 1024 ]; then
			adjusted=1024
		fi
		warn "请求内存 (${req_mem}MB) 超过 Docker 可用 (${DOCKER_MEM_MB}MB)，自动调整为 ${adjusted}MB"
		req_mem="$adjusted"
	fi

	CLUSTER_CPUS="$req_cpus"
	CLUSTER_MEM_MB="$req_mem"
}

step "Step 1/7  检查前置依赖"

MISSING=()
for cmd in docker kubectl minikube; do
	if ! command -v "$cmd" >/dev/null 2>&1; then
		MISSING+=("$cmd")
	fi
done

if [ ${#MISSING[@]} -gt 0 ]; then
	error "缺少以下依赖，请先安装: ${MISSING[*]}"
	echo ""
	echo "  macOS 一键安装:"
	echo "    brew install --cask docker"
	echo "    brew install kubectl minikube"
	echo ""
	exit 1
fi

if ! docker info >/dev/null 2>&1; then
	error "Docker daemon 未运行，请先启动 Docker Desktop"
	exit 1
fi

detect_docker_resources
resolve_cluster_resources

info "依赖检查通过 (docker, kubectl, minikube)"
info "Docker 资源: ${DOCKER_CPUS} CPU / ${DOCKER_MEM_MB}MB RAM"
info "Minikube 将使用: ${CLUSTER_CPUS} CPU / ${CLUSTER_MEM_MB}MB RAM / ${DISK_SIZE} disk"

step "Step 2/7  启动 Minikube 集群"

if minikube status -p "${CLUSTER_NAME}" >/dev/null 2>&1; then
	info "Minikube 集群 '${CLUSTER_NAME}' 已在运行"
else
	info "Minikube 集群 '${CLUSTER_NAME}' 未运行，正在启动 (${CLUSTER_CPUS} CPU / ${CLUSTER_MEM_MB}MB RAM / ${DISK_SIZE} disk)..."
	minikube start -p "${CLUSTER_NAME}" \
		--cpus="${CLUSTER_CPUS}" --memory="${CLUSTER_MEM_MB}" --disk-size="${DISK_SIZE}" \
		--driver=docker \
		--addons=metrics-server \
		--wait=all
fi

kubectl config use-context "${CLUSTER_NAME}" >/dev/null 2>&1
MINIKUBE_IP=$(minikube ip -p "${CLUSTER_NAME}")
info "集群就绪，Minikube IP: ${MINIKUBE_IP}"

step "Step 3/7  配置 Docker 指向 Minikube 内部 daemon"
eval "$(minikube docker-env -p "${CLUSTER_NAME}")"
info "Docker 已切换到 Minikube 内部 daemon (镜像直接构建到集群中，无需 push/load)"

step "Step 4/7  构建 Docker 镜像"

cd "${PROJECT_ROOT}"
info "构建 argus-gateway:dev ..."
docker build --target gateway \
	-t argus-gateway:dev \
	--build-arg "VERSION=dev-${CLUSTER_NAME}" \
	-q .

info "构建 argus-controller:dev ..."
docker build --target controller \
	-t argus-controller:dev \
	--build-arg "VERSION=dev-${CLUSTER_NAME}" \
	-q .

info "镜像构建完成"
docker images | grep -E 'argus-(gateway|controller)' | head -5 | while read -r line; do
	echo "         $line"
done

step "Step 5/7  部署 Argus 到 K8s"

kubectl apply -k k8s/overlays/dev

echo ""
info "资源已提交，等待 Deployment 就绪..."

step "Step 6/7  等待 Pod 就绪 (超时 ${WAIT_TIMEOUT}s)"

if ! kubectl wait --for=condition=Available --timeout="${WAIT_TIMEOUT}s" \
	deployment/argus-gateway deployment/argus-controller \
	-n "${NAMESPACE}" 2>&1; then
	warn "等待超时，当前 Pod 状态:"
	kubectl get pods -n "${NAMESPACE}" -o wide
	echo ""
	warn "可执行 'make dev-logs' 查看日志排查问题"
fi

echo ""
kubectl get pods -n "${NAMESPACE}" -o wide 2>/dev/null || true

step "Step 7/7  启动端口转发"

pkill -f "kubectl.*port-forward.*${NAMESPACE}/svc/argus" 2>/dev/null || true
sleep 1

kubectl port-forward -n "${NAMESPACE}" svc/argus-gateway "${GATEWAY_PORT}:8443" >/dev/null 2>&1 &
PF_GATEWAY=$!
kubectl port-forward -n "${NAMESPACE}" svc/argus-gateway "${GATEWAY_METRICS_PORT}:9090" >/dev/null 2>&1 &
PF_GATEWAY_M=$!
kubectl port-forward -n "${NAMESPACE}" svc/argus-controller "${CONTROLLER_GRPC_PORT}:8444" >/dev/null 2>&1 &
PF_CTRL=$!
kubectl port-forward -n "${NAMESPACE}" svc/argus-controller "${CONTROLLER_METRICS_PORT}:9090" >/dev/null 2>&1 &
PF_CTRL_M=$!

sleep 2
echo "${PF_GATEWAY} ${PF_GATEWAY_M} ${PF_CTRL} ${PF_CTRL_M}" > "${STATE_DIR}/port-forward.pids"
echo "${CLUSTER_NAME}" > "${STATE_DIR}/cluster-name"
echo "${NAMESPACE}" > "${STATE_DIR}/namespace"

echo ""
echo -e "${GREEN}${BOLD}╔══════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}${BOLD}║            Argus (观枢) 开发环境已启动!             ║${NC}"
echo -e "${GREEN}${BOLD}╠══════════════════════════════════════════════════════╣${NC}"
printf "${GREEN}${BOLD}║${NC} ${BOLD}Gateway Proxy:${NC}     ${CYAN}https://localhost:%d${NC}\n" "${GATEWAY_PORT}"
printf "${GREEN}${BOLD}║${NC} ${BOLD}Gateway Metrics:${NC}   ${CYAN}http://localhost:%d/metrics${NC}\n" "${GATEWAY_METRICS_PORT}"
printf "${GREEN}${BOLD}║${NC} ${BOLD}Gateway Health:${NC}    ${CYAN}http://localhost:%d/healthz${NC}\n" "${GATEWAY_METRICS_PORT}"
printf "${GREEN}${BOLD}║${NC} ${BOLD}Controller gRPC:${NC}   ${CYAN}localhost:%d${NC}\n" "${CONTROLLER_GRPC_PORT}"
printf "${GREEN}${BOLD}║${NC} ${BOLD}Controller Metrics:${NC}${CYAN} http://localhost:%d/metrics${NC}\n" "${CONTROLLER_METRICS_PORT}"
echo -e "${GREEN}${BOLD}╠══════════════════════════════════════════════════════╣${NC}"
echo -e "${GREEN}${BOLD}║${NC} ${BOLD}Minikube Dashboard:${NC} ${CYAN}minikube dashboard -p ${CLUSTER_NAME}${NC}"
echo -e "${GREEN}${BOLD}║${NC} ${BOLD}常用命令:${NC}"
echo -e "${GREEN}${BOLD}║${NC}   make dev-status   查看 Pod/Service 状态"
echo -e "${GREEN}${BOLD}║${NC}   make dev-logs     实时查看所有 Pod 日志"
echo -e "${GREEN}${BOLD}║${NC}   make dev-down     停止环境 (删除集群)"
echo -e "${GREEN}${BOLD}╚══════════════════════════════════════════════════════╝${NC}"
echo ""
info "端口转发运行中，按 Ctrl+C 停止端口转发（集群保留）"

wait
