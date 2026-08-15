# assets/

纯静态资源：logo、架构图（只收 SVG）。PNG/JPG 一律转 SVG 再放，不然 repo 会涨得比工资还快。

子目录：
- `logo/`：argus-logo（横版/方形、深色背景/浅色背景各一份）。
- `diagrams/`：README 用的图，源文件是 drawio.svg 也可以。
- `brand/`：字体授权 + 配色表（合规要用）。

提交 PR 前跑一下：`ls -lh assets/**/*`，单个 SVG 超过 1MB 先问一下："这张真的需要吗？"
