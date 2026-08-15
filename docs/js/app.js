/* Argus Docs 着陆页逻辑（React 18 + antd v5 UMD）
 * 展示：
 *   - 快速导航（架构文档、流量方案、TLS、运行模式矩阵、Pod 身份、Helm values）
 *   - 4 个 KPI 卡片（组件数量/CRD 数量 / 支持引擎 / 支持云厂商模块）
 *   - 运行模式 2×2（mode/failureMode）矩阵 Table + 严重度 Tag
 *   - 下载 CRD / Deploy / Proto 一组 Action（走本地相对路径，不依赖网络）
 *   - 全局中文 ConfigProvider + dayjs zh-cn locale
 */
(function (global) {
  'use strict';

  const h = (global.React && global.React.createElement) || function () { return null; };
  const e = h;

  function useAntd() {
    return global.antd;
  }
  function useIcons() {
    return global.icons;
  }

  function to(n) { return n + 1; } // unused guard, removed by minifier if any
  void to;

  // ---- Data ----
  const NAV_LINKS = [
    { key: 'arch', label: '架构设计', href: 'architecture.md' },
    { key: 'ti',   label: '流量引流', href: 'traffic-interception.md' },
    { key: 'pod',  label: 'Pod 身份', href: 'pod-identity.md' },
    { key: 'tls',  label: 'TLS 解密', href: 'tls-design.md' },
    { key: 'run',  label: '运行模式矩阵', href: 'runmodes-failure.md' },
    { key: 'helm', label: 'Helm Values', href: '../deploy/helm/argus/values.yaml' }
  ];

  const KPI = [
    { key: 'proto', label: 'gRPC 契约数', value: 5,
      tip: 'policy / detection / event / health / identity 五类 proto 服务' },
    { key: 'crd', label: 'K8s CRD 数', value: 2,
      tip: 'ArgusSecurityPolicy + LLMProvider 两类 CRD' },
    { key: 'engines', label: '支持引流引擎', value: '2+',
      tip: 'Cilium eBPF（首选）/ iptables TPROXY（兜底）' },
    { key: 'clouds', label: 'Terraform 云厂商模块', value: 3,
      tip: 'AWS EKS / GCP GKE / Azure AKS' }
  ];

  const MATRIX = [
    // 摘取 runmodes-failure.md 里最常问的 8 行 × 4 列，完整表跳转 runmodes-failure.md
    { scenario: 'controller Policy gRPC 断连 / 快照过期',
      mo_open: { tag: 'success', t: 'ALLOW' }, mo_closed: { tag: 'success', t: 'ALLOW' },
      eo_open: { tag: 'warning', t: 'DEGRADED_ALLOW' }, eo_closed: { tag: 'error',   t: 'DEGRADED_BLOCK' },
      reason: 'controller_unreachable' },
    { scenario: '单一检测器超时（rules/heuristic/encoding）',
      mo_open: { tag: 'success', t: 'ALLOW' }, mo_closed: { tag: 'success', t: 'ALLOW' },
      eo_open: { tag: 'warning', t: 'DEGRADED_ALLOW' }, eo_closed: { tag: 'error', t: 'DEGRADED_BLOCK' },
      reason: 'detector_timeout' },
    { scenario: 'CA Secret 读取失败（B 模式）',
      mo_open: { tag: 'warning', t: '降级 A 模式' }, mo_closed: { tag: 'error', t: '阻断' },
      eo_open: { tag: 'warning', t: '降级 A 模式' }, eo_closed: { tag: 'error', t: '阻断' },
      reason: 'ca_secret_missing' },
    { scenario: '事件连续 3 次上报失败（仅审计失败）',
      mo_open: { tag: 'success', t: '按 monitor 放行' }, mo_closed: { tag: 'success', t: '按 monitor 放行' },
      eo_open: { tag: 'processing', t: '按风险分决策' }, eo_closed: { tag: 'processing', t: '按风险分决策' },
      reason: 'event_sink_degraded（绝不改变决策）' }
  ];

  // ---- Helpers ----
  function TagColor(kind) {
    // antd Tag color mapping 名称，避免硬编码 8 种不同色值
    switch (kind) {
      case 'success': return 'green';
      case 'warning': return 'gold';
      case 'error': return 'red';
      case 'processing': return 'blue';
      default: return 'default';
    }
  }

  function downloadFromRelative(relPath, fallbackName) {
    // 从相对路径下载：如果是 .md / .yaml，浏览器直接打开时也能右键另存；
    // 这里用 <a download> 模拟触发一次下载。
    const a = document.createElement('a');
    a.href = relPath;
    a.rel = 'noopener noreferrer';
    if (fallbackName) a.setAttribute('download', fallbackName);
    document.body.appendChild(a);
    try { a.click(); } finally { document.body.removeChild(a); }
  }

  // ---- Sections ----
  function Header() {
    const antd = useAntd();
    const icons = useIcons();
    const Space = antd.Space;
    const Button = antd.Button;
    return e('header', { className: 'app-header' },
      e('div', { className: 'app-brand' },
        e('div', { className: 'brand-dot' }),
        e('span', null, 'Argus（观枢）文档中心')
      ),
      e('nav', { className: 'app-nav' },
        NAV_LINKS.map(l => e('a', { key: l.key, className: 'nav-link', href: l.href }, l.label))
      ),
      e(Space, { size: 8 },
        e(Button, { type: 'default', size: 'small', icon: e(icons.GithubOutlined),
          onClick: () => window.open('https://github.com/ctkqiang/argus', '_blank', 'noopener') }, 'GitHub'),
        e(Button, { type: 'primary', size: 'small', icon: e(icons.CloudDownloadOutlined),
          onClick: () => downloadFromRelative('../deploy/helm/argus/Chart.yaml') }, 'Helm Chart')
      )
    );
  }

  function Hero() {
    const antd = useAntd();
    const icons = useIcons();
    const Button = antd.Button;
    const Tag = antd.Tag;
    return e('section', { className: 'hero' },
      e('div', { className: 'hero-inner' },
        e('div', { className: 'pill' }, e(icons.RadarChartOutlined, null), 'Kubernetes 原生 AI 出口安全网关'),
        e('h1', null,
          '为你的大模型出向流量，',
          e('span', { className: 'gradient' }, '做一次专业的安全观枢')
        ),
        e('p', { className: 'lead' },
          '以零业务改造的透明出口引流方式，把集群里所有 LLM 出站请求统一过检测流水线、统一记审计事件、',
          '统一做 TLS 可视性与 CA 信任管理。业务 Pod 不装 sidecar、不改 SDK、不改 BASE_URL。'
        ),
        e('div', { className: 'hero-actions' },
          e(Button, { type: 'primary', size: 'large', icon: e(icons.RocketOutlined),
            onClick: () => { window.location.href = 'architecture.md'; } }, '先看架构设计'),
          e(Button, { size: 'large', icon: e(icons.DeploymentUnitOutlined),
            onClick: () => downloadFromRelative('../deploy/helm/argus/values.yaml', 'values.yaml') }, '下载 values.yaml'),
          e(Tag, { color: 'blue', style: { alignSelf: 'center' } }, 'antd v5 · React 18 · dayjs zh-cn')
        )
      )
    );
  }

  function KpiCards() {
    const antd = useAntd();
    const icons = useIcons();
    const { Card, Tooltip, Statistic, Row, Col } = antd;
    const ICONS = {
      proto: icons.ApiOutlined,
      crd: icons.ClusterOutlined,
      engines: icons.ForkOutlined,
      clouds: icons.CloudServerOutlined
    };
    return e('div', { className: 'kpi-grid' }, KPI.map(k =>
      e(Tooltip, { key: k.key, title: k.tip },
        e(Card, null,
          e(Row, { gutter: 12, align: 'middle' },
            e(Col, { span: 6 }, e(ICONS[k.key], { style: { fontSize: 28, color: '#1677ff' } })),
            e(Col, { span: 18 }, e(Statistic, { title: k.label, value: k.value }))
          )
        )
      )
    ));
  }

  function QuickLinks() {
    const antd = useAntd();
    const icons = useIcons();
    const { Card, List, Tag } = antd;
    const tagColor = {
      arch: 'blue', ti: 'purple', pod: 'cyan', tls: 'geekblue', run: 'magenta', helm: 'volcano'
    };
    const data = NAV_LINKS.map(n => ({
      title: n.label,
      desc: n.href,
      tag: tagColor[n.key] || 'default',
      href: n.href
    }));
    return e(Card, { title: '设计文档导航', size: 'default' },
      e(List, {
        itemLayout: 'horizontal',
        dataSource: data,
        renderItem: function (item) {
          return e(List.Item, { actions: [
            e(Tag, { color: item.tag, key: 't' }, item.href.split('/').pop())
          ] },
            e(List.Item.Meta, {
              avatar: e(icons.FileTextOutlined, { style: { color: '#1677ff' } }),
              title: e('a', { href: item.href, rel: 'noopener' }, item.title),
              description: '跳转到仓库中的具体文档或配置（浏览器可直接预览或另存）'
            })
          );
        }
      })
    );
  }

  function MatrixTable() {
    const antd = useAntd();
    const { Table, Tag, Button, Space, Tooltip } = antd;
    const icons = useIcons();
    const columns = [
      { title: '故障场景', dataIndex: 'scenario', key: 's', width: 320 },
      { title: 'monitor + fail-open',  dataIndex: 'mo_open',   key: 'mo_open',
        render: v => e(Tag, { color: TagColor(v.tag) }, v.t) },
      { title: 'monitor + fail-closed', dataIndex: 'mo_closed', key: 'mo_closed',
        render: v => e(Tag, { color: TagColor(v.tag) }, v.t) },
      { title: 'enforce + fail-open',  dataIndex: 'eo_open',   key: 'eo_open',
        render: v => e(Tag, { color: TagColor(v.tag) }, v.t) },
      { title: 'enforce + fail-closed', dataIndex: 'eo_closed', key: 'eo_closed',
        render: v => e(Tag, { color: TagColor(v.tag) }, v.t) },
      { title: 'AIEvent.failure_reason', dataIndex: 'reason', key: 'r' }
    ];
    return e(Table, {
      size: 'middle',
      pagination: false,
      columns: columns,
      dataSource: MATRIX.map((m, i) => ({ key: i, ...m })),
      title: () => e(Space, { size: 12, wrap: true },
        '§12.3 故障 × 运行模式 矩阵（节选，完整 24 格见文档）',
        e(Tooltip, { title: '跳转 docs/runmodes-failure.md 查看 24 格全表 + 语义边界' },
          e(Button, { type: 'link', size: 'small', icon: e(icons.LinkOutlined),
            onClick: () => { window.location.href = 'runmodes-failure.md'; } }, '查看完整矩阵')
        )
      )
    });
  }

  function Actions() {
    const antd = useAntd();
    const icons = useIcons();
    const { Card, Button, Space, Divider, Tooltip } = antd;
    const buttons = [
      { label: '下载 values.yaml',
        icon: icons.SettingOutlined,
        onClick: () => downloadFromRelative('../deploy/helm/argus/values.yaml', 'values.yaml'),
        tip: 'Helm 默认 values（可直接改副本/镜像/阈值）' },
      { label: '下载 LLMProvider 样例',
        icon: icons.ApiOutlined,
        onClick: () => downloadFromRelative('../deploy/examples/llmprovider-openai.yaml'),
        tip: '注册 OpenAI 为合法出站点的样例 CR' },
      { label: '下载 ArgusSecurityPolicy 样例',
        icon: icons.ShieldOutlined,
        onClick: () => downloadFromRelative('../deploy/examples/argussecuritypolicy-default.yaml'),
        tip: '默认命名空间 enforce + fail-open 策略' },
      { label: '查看 Helm Chart',
        icon: icons.AppstoreOutlined,
        onClick: () => { window.location.href = '../deploy/helm/argus/Chart.yaml'; },
        tip: 'Chart metadata（版本/维护者/关键词）' }
    ];
    return e(Card, { title: '交付物下载 / 预览', size: 'default' },
      e(Space, { wrap: true, size: 10 }, buttons.map(b =>
        e(Tooltip, { key: b.label, title: b.tip },
          e(Button, { icon: e(b.icon), onClick: b.onClick }, b.label)
        )
      )),
      e(Divider, { style: { margin: '16px 0' } }),
      e('div', { style: { color: 'rgba(0,0,0,0.55)', fontSize: 13, lineHeight: 1.7 } },
        '说明：本页所有下载都走仓库相对路径，',
        '离线部署时可以把 docs/ 目录整个打包，不必担心外链缺失；',
        '仅 React/antd/dayjs/icons 走官方 CDN（unpkg 优先，失败自动回落到 jsdelivr）。'
      )
    );
  }

  function Footer() {
    const antd = useAntd();
    const { Divider } = antd;
    return e('footer', { className: 'app-footer' },
      e('div', { className: 'row' },
        'Made with care by 哪吒网络安全 · Argus 核心团队',
        e(Divider, { type: 'vertical' }),
        'License: Apache-2.0'
      )
    );
  }

  function App() {
    const antd = useAntd();
    const { ConfigProvider, App as AntApp } = antd;
    const zhCN = (antd.locales && antd.locales.zhCN) ? antd.locales.zhCN : undefined;
    return e(ConfigProvider, { locale: zhCN, theme: { token: { colorPrimary: '#1677ff', borderRadius: 8 } } },
      e(AntApp, null,
        e('div', { className: 'app-shell' },
          e(Header, null),
          e(Hero, null),
          e('main', { className: 'content' },
            e(KpiCards, null),
            e('div', { className: 'section-title', style: { marginBottom: 12 } },
              e('h2', null, '文档与样例'),
              e('span', { className: 'sub' }, '点击链接可直接跳转到仓库内文档')
            ),
            e(QuickLinks, null),
            e('div', { className: 'section-title', style: { marginTop: 36 } },
              e('h2', null, '故障 & 运行模式矩阵（节选）'),
              e('span', { className: 'sub' }, '完整 24 格请看 runmodes-failure.md')
            ),
            e(MatrixTable, null),
            e('div', { className: 'section-title', style: { marginTop: 36 } },
              e('h2', null, '快速上手：下载配置与样例'),
              e('span', { className: 'sub' }, 'Helm install 之前，先看 values.yaml')
            ),
            e(Actions, null)
          ),
          e(Footer, null)
        )
      )
    );
  }

  function renderRoot() {
    const ReactDOM = global.ReactDOM;
    const container = document.getElementById('root');
    if (!container || !ReactDOM) return;
    if (ReactDOM.createRoot) {
      ReactDOM.createRoot(container).render(h(App, null));
    } else {
      ReactDOM.render(h(App, null), container);
    }
  }

  // 导出给 index.html：ArgusDocsApp.mount()
  global.ArgusDocsApp = {
    mount: function () {
      if (global.dayjs && typeof global.dayjs.locale === 'function') {
        try { global.dayjs.locale('zh-cn'); } catch (_) { /* noop */ }
      }
      if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', renderRoot, { once: true });
      } else {
        renderRoot();
      }
    }
  };
})(typeof window !== 'undefined' ? window : this);
