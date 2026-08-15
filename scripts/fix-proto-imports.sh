#!/usr/bin/env bash
# scripts/fix-proto-imports.sh
# 快速诊断并修复 "Import 'argus/detection.proto' was not found or had errors" 问题。
# 根因：protoc 或 IDE 的 proto include root（--proto_path / includes）
#       未指向项目内的 proto/ 目录，导致相对路径 "argus/xxx.proto" 找不到文件。
# 修复动作：1) 用正确 --proto_path 原生 protoc 重新验证；2) 调用 buf build/lint/generate；
#           3) 给出 IDE 手动配置说明。
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
PROTO_DIR="${ROOT_DIR}/proto"

echo "=========================================================="
echo "[fix-proto-imports] Root:   ${ROOT_DIR}"
echo "[fix-proto-imports] Proto:  ${PROTO_DIR}"
echo "=========================================================="

echo ""
echo "== 1/4 确认 detection.proto 存在并带正确 package =="
DET="${PROTO_DIR}/argus/detection.proto"
if [ ! -f "${DET}" ]; then
  echo "ERROR: ${DET} 不存在，先恢复 detection.proto 文件。"
  exit 1
fi
echo "OK  文件存在：${DET}"
grep -E '^package |^option go_package |^message DetectionResult ' "${DET}"

echo ""
echo "== 2/4 原生 protoc 验证（--proto_path=${PROTO_DIR} 才是关键） =="
if ! command -v protoc >/dev/null 2>&1; then
  echo "SKIP: protoc 未安装，跳过原生验证。"
else
  TMP_DS="$(mktemp /tmp/argus_ds_XXXXXX)"
  # 这里 --proto_path 设为 ${PROTO_DIR} 后，import "argus/detection.proto"
  # 会拼成 ${PROTO_DIR}/argus/detection.proto，刚好命中。
  (cd "${PROTO_DIR}" && protoc --proto_path=. \
         --descriptor_set_out="${TMP_DS}" \
         --include_imports $(find argus -name '*.proto' | sort))
  echo "OK  原生 protoc 生成 descriptor 成功（无 Import was not found 报错）。"
  rm -f "${TMP_DS}"
fi

echo ""
echo "== 3/4 buf build + buf lint 官方工具链验证 =="
if ! command -v buf >/dev/null 2>&1; then
  echo "SKIP: buf 未安装，跳过 buf 工具链校验。"
else
  (cd "${PROTO_DIR}" && buf build --config buf.yaml)
  echo "OK  buf build 通过。"
  (cd "${PROTO_DIR}" && buf lint --config buf.yaml)
  echo "OK  buf lint  通过。"
fi

echo ""
echo "== 4/4 IDE 手动修复指引（仅 buf 修复仍红波浪线时需做） =="
cat <<'EOF'

   问题根源：IDE 的 Protobuf 插件默认把 workspaceRoot 当 proto include root，
            但本项目 proto 文件统一放在 proto/ 子目录内，因此 import "argus/xxx.proto"
            会去 <workspace>/argus/xxx.proto 找，自然找不到。

   VSCode 两种修复二选一（推荐方式 A）：

   A. 工作区配置 .vscode/settings.json 追加（推荐，团队统一）：
      {
        "protoc": { "options": [ "--proto_path=${workspaceRoot}/proto" ] },
        "buf":    { "inputConfig": "${workspaceRoot}/proto/buf.yaml" }
      }
      注：.vscode 目录可能被 repo 规则拦截，可粘贴到「首选项 -> 设置 -> 工作区」GUI。

   B. 插件级设置（只对本机生效）：
      设置 -> 搜索 protoc.options -> 追加 --proto_path=/你的/工程/绝对/路径/proto

   GoLand / IntelliJ：
      Preferences → Languages & Frameworks → Protocol Buffers
      → 取消 "Configure automatically"
      → 添加 Include path: <project_root>/proto

EOF

echo "=========================================================="
echo "[fix-proto-imports] DONE. 一切就绪。"
echo "=========================================================="
