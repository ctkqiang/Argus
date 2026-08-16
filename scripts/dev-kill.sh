#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
STATE_DIR="${PROJECT_ROOT}/.argus-dev"

CLUSTER_NAME="${ARGUS_CLUSTER:-argus-dev}"
if [ -f "${STATE_DIR}/cluster-name" ]; then
	CLUSTER_NAME=$(cat "${STATE_DIR}/cluster-name")
fi

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC}  $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
step()  { echo -e "\n${CYAN}${BOLD}==>${NC} ${BOLD}$*${NC}"; }

step "停止端口转发"
killed=0
if [ -f "${STATE_DIR}/port-forward.pids" ]; then
	while read -r pid; do
		if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
			kill "$pid" 2>/dev/null || true
			killed=$((killed + 1))
		fi
	done < "${STATE_DIR}/port-forward.pids"
	rm -f "${STATE_DIR}/port-forward.pids"
fi
pkill -f "kubectl.*port-forward.*argus" 2>/dev/null || true
if [ "$killed" -gt 0 ]; then
	info "已停止 ${killed} 个端口转发进程"
else
	info "没有运行中的端口转发"
fi

step "停止 Minikube 集群 '${CLUSTER_NAME}'"
if minikube status -p "${CLUSTER_NAME}" >/dev/null 2>&1; then
	minikube stop -p "${CLUSTER_NAME}"
	info "集群已停止（镜像和部署数据保留，下次 make run 秒恢复）"
else
	info "集群 '${CLUSTER_NAME}' 未运行，跳过"
fi

echo ""
info "Argus 开发环境已停止。"
info "  重新启动（保留数据）: make run"
info "  彻底销毁（释放磁盘）: make dev-down"
