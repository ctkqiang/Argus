# templates/cilium/

装了 Cilium 的集群，**推荐** 用 `CiliumEgressGatewayPolicy` 把匹配 LLM 厂商域名的出口流量重定向到 argus-gateway。

占位文件：真实模板等业务代码到 1-B 阶段再加；现在只留规则：
- 不要把非 443 / 非 LLM 厂商的流量也引进来，会把网关打穿。
- 高可用要加 `toServices: [ { name: argus-gateway, namespace: argus } ]`，不要写死 PodIP。
