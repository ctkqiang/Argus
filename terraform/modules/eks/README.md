# modules/eks

最小可用的 Argus EKS 集群：VPC（3 public / 3 private）、EKS（K8s ≥ 1.27）、
Node Group（spot + on-demand 混合）、IRSA 角色（controller/gateway 各一个 SA）。

不做的事（MVP 里不做，别硬塞进来）：
- 不帮你建 cert-manager / Cilium。Helm 层自己装。
- 不帮你建 VPC endpoint / private link。需要自己在 `envs/prod/` 里调变量。
