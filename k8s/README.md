# k8s/

这里放**纯 Kubernetes YAML**（不经过 Helm 渲染的原始清单），一般用来：

- 开发环境快速裸 apply 看组件跑起来没（先不折腾 Helm values）。
- CI 里拿一份最小 manifest 做静态检查（`kubeval` / `kubeconform`）。
- 集群紧急修复时，维护者直接改这里的补丁版本。

生产部署走 `../deploy/helm/argus/`，不要直接把这里的 YAML `apply -f k8s/` 到生产集群。

子目录：
- `base/`：所有环境公用的 Deployment/Service/SA/RBAC。
- `overlays/dev/`，`overlays/stg/`，`overlays/prod/`：kustomize 层，改镜像、改副本、改 Service type。
- `patches/`：通用 JSON / strategic merge patch（比如 resources、tolerations）。
