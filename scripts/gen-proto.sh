#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
PROTO_DIR="${ROOT_DIR}/proto"
OUTPUT_DIR="${ROOT_DIR}/pkg/pb"

echo "[gen-proto] Root directory:    ${ROOT_DIR}"
echo "[gen-proto] Proto directory:   ${PROTO_DIR}"
echo "[gen-proto] Output directory:  ${OUTPUT_DIR}"

if ! command -v buf >/dev/null 2>&1; then
  echo "[gen-proto] ERROR: 'buf' command not found."
  echo "[gen-proto] Please install: go install github.com/bufbuild/buf/cmd/buf@latest"
  exit 1
fi

if ! command -v protoc-gen-go >/dev/null 2>&1; then
  echo "[gen-proto] ERROR: 'protoc-gen-go' command not found."
  echo "[gen-proto] Please install: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
  exit 1
fi

if ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then
  echo "[gen-proto] ERROR: 'protoc-gen-go-grpc' command not found."
  echo "[gen-proto] Please install: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
  exit 1
fi

mkdir -p "${OUTPUT_DIR}"

echo "[gen-proto] Running buf lint..."
cd "${PROTO_DIR}" && buf lint

echo "[gen-proto] Running buf generate..."
cd "${PROTO_DIR}" && buf generate

echo "[gen-proto] Generation complete. Output files:"
find "${OUTPUT_DIR}" -type f -name "*.go" | sort
