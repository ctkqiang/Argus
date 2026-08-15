# deploy/

部署相关的一切放这里。分 4 块：

1. `helm/argus/`：Helm Chart（正式交付都走这个）。
2. `examples/`：开箱即用的样例 CR YAML + kind 一键 up 脚本。
3. `kustomize/`：非 Helm 的 kustomize 层（通常给 airgap 场景用）。
4. `manifests/`：Helm 渲染后的静态清单（CI snapshot，不要手工改）。

怎么选：
- 正常部署：`helm install argus ./helm/argus -n argus --create-namespace`
- 本地玩：`bash examples/dev/kind-up.sh` 一键起来，然后 `kubectl apply -f examples/*.yaml`
- airgap / 强管控：kustomize build `kustomize/overlays/prod`，再推给 Argo CD
