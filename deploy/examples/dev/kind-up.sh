#!/usr/bin/env bash
# kind + Cilium + cert-manager 一键搭本地 Argus 测试集群。
# 前置：kind, kubectl, helm 都在 PATH 里；docker 已启。
set -euo pipefail

CLUSTER="${ARGUS_KIND_CLUSTER:-argus-dev}"
CILIUM_VERSION="${ARGUS_CILIUM_VERSION:-1.15.7}"
CERT_MANAGER_VERSION="${ARGUS_CERT_MANAGER_VERSION:-v1.15.3}"

if ! kind get clusters | grep -q "^${CLUSTER}\$"; then
  kind create cluster --name "${CLUSTER}"
fi

kubectl cluster-info --context "kind-${CLUSTER}"
helm repo add cilium https://helm.cilium.io/ || true
helm repo add jetstack https://charts.jetstack.io || true
helm repo update

helm upgrade --install cilium cilium/cilium   --version "${CILIUM_VERSION}"   --namespace kube-system   --set egressGateway.enabled=true

helm upgrade --install cert-manager jetstack/cert-manager   --version "${CERT_MANAGER_VERSION}"   --namespace cert-manager --create-namespace   --set installCRDs=true

echo "Done. Next: helm install argus ./deploy/helm/argus -n argus --create-namespace"
