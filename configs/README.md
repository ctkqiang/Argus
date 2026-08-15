# configs/

放两种东西：
1. `*.example` 配置模板：告诉新同事启动 gateway/controller 需要填哪些字段。
2. `*.skeleton`：给 CI / staging 用的最小可行配置（不含任何真实 secrets）。

**禁止**：
- 任何真实 sk-token / kubeconfig / tls.key 放进来。
- 把环境变量的值硬编码进 YAML（走 env/var substitution 或 k8s Secret）。
