#!/usr/bin/env python3
"""Convert docs/*.md to standalone HTML pages with Argus Ant Design theme.

Usage: python3 scripts/md2html.py
"""

import markdown
from pathlib import Path

DOCS_DIR = Path(__file__).resolve().parent.parent / "docs"

NAV_LINKS = [
    ("index.html",            "首页"),
    ("architecture.html",     "架构设计"),
    ("traffic-interception.html", "流量引流"),
    ("pod-identity.html",     "Pod 身份"),
    ("tls-design.html",       "TLS 解密"),
    ("runmodes-failure.html", "运行模式矩阵"),
    ("helm-values.html",      "Helm Values"),
]

def nav_html(active: str) -> str:
    links = []
    for href, label in NAV_LINKS:
        cls = "nav-link active" if href == active else "nav-link"
        links.append(f'<a class="{cls}" href="{href}">{label}</a>')
    return "\n".join(links)

TEMPLATE = """<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <meta name="description" content="Argus（观枢）文档中心 — {title}" />
  <title>{title} | Argus（观枢）</title>
  <link rel="preconnect" href="https://unpkg.com" crossorigin />
  <link rel="dns-prefetch" href="https://unpkg.com" />
  <link rel="stylesheet" href="https://unpkg.com/antd@5.21.4/dist/reset.css" crossorigin="anonymous" referrerpolicy="no-referrer-when-downgrade" />
  <link rel="stylesheet" href="./css/styles.css" />
  <link rel="stylesheet" href="./css/doc-page.css" />
</head>
<body>
  <div class="app-shell">
    <header class="app-header">
      <a class="app-brand" href="index.html">
        <div class="brand-dot"></div>
        <span>Argus（观枢）文档中心</span>
      </a>
      <nav class="app-nav">
        {nav}
      </nav>
      <div class="header-actions">
        <a class="btn btn-default btn-sm" href="https://github.com/ctkqiang/argus" target="_blank" rel="noopener">GitHub</a>
        <a class="btn btn-default btn-sm" href="https://gitcode.com/ctkqiang_sr/Argus" target="_blank" rel="noopener">GitCode</a>
      </div>
    </header>
    <main class="doc-content">
      {body}
    </main>
    <footer class="app-footer">
      <div class="footer-row">
        <span>Made with care by 哪吒网络安全 - Argus 核心团队</span>
        <span class="divider">|</span>
        <span>License: Apache-2.0</span>
      </div>
    </footer>
  </div>
</body>
</html>
"""

def md_to_html(md_path: Path) -> str:
    text = md_path.read_text(encoding="utf-8")
    md = markdown.Markdown(extensions=[
        "tables",
        "fenced_code",
        "toc",
        "nl2br",
        "sane_lists",
    ])
    body = md.convert(text)

    title_map = {
        "architecture":          "架构设计",
        "traffic-interception":  "流量引流",
        "pod-identity":          "Pod 身份溯源",
        "tls-design":            "TLS 解密设计",
        "runmodes-failure":      "运行模式矩阵",
        "helm-values":           "Helm Values 配置参考",
        "README":                "文档索引",
    }
    stem = md_path.stem
    title = title_map.get(stem, stem)
    active = f"{stem}.html"

    return TEMPLATE.format(
        title=title,
        nav=nav_html(active),
        body=body,
    )

def main() -> None:
    md_files = sorted(DOCS_DIR.glob("*.md"))
    if not md_files:
        print("No .md files found in docs/")
        return

    for md_path in md_files:
        html_path = md_path.with_suffix(".html")
        html_content = md_to_html(md_path)
        html_path.write_text(html_content, encoding="utf-8")
        print(f"  {md_path.name} -> {html_path.name}")

    print(f"\nDone: {len(md_files)} files converted.")

if __name__ == "__main__":
    main()
