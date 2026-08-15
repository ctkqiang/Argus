# terraform/

基础设施即代码。三条原则（对齐 `.trae/rules/go-coding-standards.md` §10.2）：

1. **Secrets 零进代码**：所有 token / 凭据走 Terraform 工作区环境变量 / cloud secret manager，
   不许写进 `*.tfvars` 或 `*.tf`。
2. **状态别本地存**：所有环境 `envs/*` 都显式 `backend "s3"` / `azurerm` / `gcs`，
   对应配置放 `backends/<env>.hcl`（backend 配置不能插值，所以单独写）。
3. **可重复**：每个模块 README 写明输入/输出/前置依赖；`terraform fmt` 必过。

目录：
- `modules/eks/`：AWS EKS + VPC + IRSA。
- `modules/gke/`：GCP GKE + VPC + Workload Identity。
- `modules/aks/`：Azure AKS + VNet + Workload Identity。
- `modules/argus-helm/`：用 Helm provider 把 `../deploy/helm/argus` 装到集群里。
- `envs/{dev,stg,prod}/`：三套环境根模块，分别调用上面的云厂商模块 + argus-helm。
- `backends/`：各环境后端配置（S3/GCS/Azurerm）。
