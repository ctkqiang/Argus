# modules/

每个模块的 README 至少写这几块：
- 作用（What / Why / When not to use）
- 前置条件（需要的云权限、已有的 VPC id、是否要装 cert-manager CRD）
- 输入变量表：name / type / default / description
- 输出变量表：name / description
- 最小可用 example：`terraform apply` 一条命令能跑

新增模块前先更新 `.trae/specs/argus-guanshu/tasks.md`。
