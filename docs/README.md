# docs/

本目录只放**长期稳定、review 过**的设计文档。临时想法、会议笔记丢 issues / Notion，别混进来。

对应 checklist §3 的 5 份文档：

1. `architecture.md`：逻辑拓扑 + 报文流转路径。
2. `traffic-interception.md`：4 种透明出口引流方案对比 + MVP 首选/兜底安装步骤。
3. `pod-identity.md`：身份溯源链路图 + 各场景准确性矩阵。
4. `tls-design.md`：被动观测 / TLS 解密两种模式、CA 根证书要求、故障降级表。
5. `runmodes-failure.md`：monitor/enforce × fail-open/fail-closed + §12.3 故障矩阵 24 格。

文档规范：
- 所有 "TODO（MVP 之后做）" 都显式写出来，别假装已经做完。
- 表格必须有"前提假设"一列，不然就是空话。
- 引用代码路径用相对路径，比如 `../internal/gateway/tls/`。
