# modules/argus-helm

把 `../../deploy/helm/argus` 这个本地 Chart 用 Terraform Helm provider 装到目标集群里。

输入：
- `kubeconfig`（或 provider helm 的 kubernetes block 注入）
- `values`：覆盖 `deploy/helm/argus/values.yaml` 的 map
- `release_name` / `namespace`

好处：基础设施和 Argus Chart 都用同一个 Terraform plan 一起审。
