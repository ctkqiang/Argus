# deploy/examples/

样例文件，只当"起点"，不要直接上生产。

- `llmprovider-openai.yaml`：注册 OpenAI 为合法出站点。
- `argussecuritypolicy-default.yaml`：default 命名空间默认策略（rules/heuristic/encoding 开，semantic 关）。
- `dev/kind-up.sh`：本地 kind + Cilium + cert-manager 一键搭测试集群（需要 docker）。

生产前必须改三件事：
1. `auth_secret_ref.name` 引用你们自己的 Secret，不要复制 example 里的 name。
2. `thresholds.risk.*` 按你们的误报/漏报调，不要默认 0.3/0.6/0.85。
3. `scope.namespaces` 别留空，别一次性把全集群包进去滚挂业务。
