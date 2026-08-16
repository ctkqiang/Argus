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
if [ -f "${STATE_DIR}/port-forward.pids" ]; then
        while read -r pid; do
                [ -n "$pid" ] && kill "$pid" 2>/dev/null || true
        done < "${STATE_DIR}/port-forward.pids"
        rm -f "${STATE_DIR}/port-forward.pids"
        info "端口转发已停止"
else
        info "没有运行中的端口转发"
fi

pkill -f "kubectl.*port-forward.*argus" 2>/dev/null || true

step "删除 Minikube 集群 '${CLUSTER_NAME}'"
if minikube status -p "${CLUSTER_NAME}" >/dev/null 2>&1; then
        minikube delete -p "${CLUSTER_NAME}"
        info "集群 '${CLUSTER_NAME}' 已删除"
else
        info "集群 '${CLUSTER_NAME}' 不存在，跳过"
fi

rm -rf "${STATE_DIR}"
info "状态文件已清理"

echo ""
info "Argus 开发环境已完全销毁。再次启动: make run"
