# deploy/examples/dev

本地开发集群脚本。跑之前先读 `kind-up.sh`，它会装 Cilium/cert-manager 这两个重东西。

常见坑：
- M1/M2 mac：docker 的 VM 内存至少给 6Gi，不然 kind 起 Cilium 会 OOM。
- 公司内网 proxy：export HTTPS_PROXY 之后，helm 拉 chart 记得把域名加到 NO_PROXY。
