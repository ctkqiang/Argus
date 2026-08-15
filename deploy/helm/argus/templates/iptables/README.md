# templates/iptables/

没装 Cilium 的兜底方案：节点上装 tproxy DaemonSet（`argus-system` 命名空间），
把出向 443 流量 TPROXY 到 gateway。

**硬限制一定要写在设计里**：
- 只能处理 TCP（UDP DNS 不在这层搞）。
- 节点已经用了别的 tproxy（比如 service-mesh 出向）会冲突，二选一。
- 回滚：删除 DaemonSet 后逐条删除 PREROUTING/OUTPUT 里加的链，不要直接 iptables -F。
