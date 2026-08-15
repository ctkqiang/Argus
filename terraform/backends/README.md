# backends/

每个环境一份 `*.hcl`，给 `terraform init -backend-config=backends/<env>.hcl` 用。

backend 配置不能做变量插值，所以单独写。格式建议：
- S3（AWS）：bucket / key / region / dynamodb_table（锁）。
- GCS：bucket / prefix。
- Azure：resource_group_name / storage_account_name / container_name / key。

本文件里所有 `.hcl.example` 都只是**模板**，不要把真实 bucket 名、account 名提交进来；
真实 backend hcl 放同事本地或用 Atlantis 环境变量注入，加进 `.gitignore`。
