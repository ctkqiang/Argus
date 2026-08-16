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

# ctkqiang 唯一作者身份；所有 canonical / OG / Twitter / JSON-LD 均绑定此身份
GH_AUTHOR_URL = "https://github.com/ctkqiang"
GITCODE_URL = "https://gitcode.com/ctkqiang_sr/Argus"
OG_IMAGE = "https://raw.githubusercontent.com/ctkqiang/argus/main/docs/images/architecture-overview.svg"
LICENSE_URL = "https://www.apache.org/licenses/LICENSE-2.0"
DATE_PUBLISHED = "2026-08-16"
DATE_MODIFIED = "2026-08-16"

# 每页 SEO 参数（description / keywords / canonical_path）
SEO_PARAMS = {
    "README": {
        "description": "Argus（观枢）文档索引 — Kubernetes 原生 AI 出口安全网关全文档导航，含架构/引流/Pod 身份/TLS/运行模式矩阵",
        "keywords": "Argus,观枢,Kubernetes,LLM 安全,AI 出口网关,文档索引,ctkqiang",
        "canonical_path": "docs/README.html",
    },
    "architecture": {
        "description": "Argus（观枢）架构设计 — 逻辑拓扑、组件职责矩阵、13 步请求流转、K8s 部署拓扑、数据一致性",
        "keywords": "Argus 架构,观枢,Kubernetes 网关,LLM 检测流水线,Argus gRPC,ctkqiang",
        "canonical_path": "docs/architecture.html",
    },
    "traffic-interception": {
        "description": "Argus 透明出口引流四方案对比 — Cilium eBPF 首选 / iptables TPROXY 兜底 / Istio 排除 / VPC 特例",
        "keywords": "Cilium eBPF,iptables TPROXY,K8s 出口引流,LLM 透明代理,Argus,ctkqiang",
        "canonical_path": "docs/traffic-interception.html",
    },
    "pod-identity": {
        "description": "Argus Pod 身份溯源链路 — gateway 零 RBAC + controller 唯一真源 + 5 场景准确性矩阵 + 降级策略",
        "keywords": "Pod 身份溯源,K8s RBAC,informer cache,Argus Pod Identity,ctkqiang",
        "canonical_path": "docs/pod-identity.html",
    },
    "tls-design": {
        "description": "Argus TLS 解密设计 — 被动观测 vs TLS 终止检测 + CA 信任模型 + 证书轮换 + 6 条故障降级表",
        "keywords": "TLS 解密,MutatingWebhook,CA 信任链,动态签叶证,Argus TLS,ctkqiang",
        "canonical_path": "docs/tls-design.html",
    },
    "runmodes-failure": {
        "description": "Argus 运行模式矩阵 — monitor/enforce × fail-open/closed 24 格 + 8 故障场景决策表 + 运维铁律",
        "keywords": "Argus 运行模式,fail-open,fail-closed,故障降级,24 格矩阵,ctkqiang",
        "canonical_path": "docs/runmodes-failure.html",
    },
}

def seo_block(title: str, description: str, keywords: str, canonical: str) -> str:
    """生成 SEO metadata 块：含 author/copyright/robots/keywords/canonical/OG/Twitter/JSON-LD。

    所有标签均绑定 ctkqiang 为唯一作者实体（@id 指向 https://github.com/ctkqiang），
    与 sameAs 中的 GitCode 仓库互为镜像身份。
    """
    return f"""  <!-- ===== SEO Meta（ctkqiang 专属，唯一作者身份） ===== -->
  <meta name="author" content="ctkqiang" />
  <meta name="copyright" content="ctkqiang" />
  <meta name="generator" content="Argus Docs Builder by ctkqiang" />
  <meta name="robots" content="index, follow, max-image-preview:large, max-snippet:-1, max-video-preview:-1" />
  <meta name="keywords" content="{keywords}" />
  <link rel="canonical" href="{canonical}" />

  <!-- ===== Open Graph（社交分享卡片，仅 ctkqiang 域下有效） ===== -->
  <meta property="og:type" content="article" />
  <meta property="og:site_name" content="Argus by ctkqiang" />
  <meta property="og:locale" content="zh_CN" />
  <meta property="og:title" content="{title}" />
  <meta property="og:description" content="{description}" />
  <meta property="og:url" content="{canonical}" />
  <meta property="og:image" content="{OG_IMAGE}" />
  <meta property="og:image:alt" content="Argus 系统架构总览图" />

  <!-- ===== Twitter Card（ctkqiang） ===== -->
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content="{title}" />
  <meta name="twitter:description" content="{description}" />
  <meta name="twitter:image" content="{OG_IMAGE}" />
  <meta name="twitter:image:alt" content="Argus 系统架构总览图" />
  <meta name="twitter:creator" content="ctkqiang" />

  <!-- ===== 结构化数据 JSON-LD（ctkqiang 作者实体） ===== -->
  <script type="application/ld+json">
  {{
    "@context": "https://schema.org",
    "@type": "TechArticle",
    "headline": "{title}",
    "description": "{description}",
    "author": {{
      "@type": "Person",
      "@id": "{GH_AUTHOR_URL}",
      "name": "ctkqiang",
      "url": "{GH_AUTHOR_URL}",
      "sameAs": [
        "{GH_AUTHOR_URL}",
        "{GITCODE_URL}"
      ]
    }},
    "publisher": {{
      "@type": "Organization",
      "@id": "{GH_AUTHOR_URL}",
      "name": "ctkqiang",
      "url": "{GH_AUTHOR_URL}"
    }},
    "mainEntityOfPage": {{
      "@type": "WebPage",
      "@id": "{canonical}"
    }},
    "image": "{OG_IMAGE}",
    "datePublished": "{DATE_PUBLISHED}",
    "dateModified": "{DATE_MODIFIED}",
    "inLanguage": "zh-CN",
    "license": "{LICENSE_URL}"
  }}
  </script>
"""

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
  {seo}
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

    seo = SEO_PARAMS.get(stem, SEO_PARAMS["README"])
    canonical = f"https://github.com/ctkqiang/argus/blob/main/{seo['canonical_path']}"
    seo_meta = seo_block(title, seo["description"], seo["keywords"], canonical)

    return TEMPLATE.format(
        title=title,
        nav=nav_html(active),
        body=body,
        seo=seo_meta,
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
