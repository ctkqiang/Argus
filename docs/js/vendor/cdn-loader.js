/* 统一维护 Argus Docs 的 CDN 依赖源：
 *   1) 默认 unpkg（Ant Design 文档推荐），失败时走 jsdelivr（与 antd 官网上的『国内镜像』思路一致）
 *   2) 仅在 window 上没有对应全局变量时才追加脚本，避免重复加载
 *   3) 所有资源版本号 pin 死在一个对象上，方便后续升级一处改
 */
(function (global) {
  'use strict';

  const VERSIONS = {
    react: '18.3.1',
    antd: '5.21.4',
    icons: '5.6.1',
    dayjs: '1.11.13'
  };

  const PRIMARY = 'https://unpkg.com';
  const FALLBACK = 'https://cdn.jsdelivr.net/npm';

    const ASSETS = [
    {
      kind: "css",
      name: "antd reset.css",
      // // no global registration required (side-effect only),
      primary:   "https://unpkg.com/antd@5.21.4/dist/reset.css",
      fallback:  "https://cdn.jsdelivr.net/npm/antd@5.21.4/dist/reset.css",
      integrity: "sha384-Rh5FRX7P6eIAOLQJXihyUaEa7SElWzcO7ZQk3pb6YhVJVfBq1m7QAZXxSy+lIAuJ",  // SRI sha384，与 docs/js/resource-manifest.json 保持一致
    },
    {
      kind: "js",
      name: "react",
      global: "React",
      primary:   "https://unpkg.com/react@18.3.1/umd/react.production.min.js",
      fallback:  "https://cdn.jsdelivr.net/npm/react@18.3.1/umd/react.production.min.js",
      integrity: "sha384-DGyLxAyjq0f9SPpVevD6IgztCFlnMF6oW/XQGmfe+IsZ8TqEiDrcHkMLKI6fiB/Z",  // SRI sha384，与 docs/js/resource-manifest.json 保持一致
    },
    {
      kind: "js",
      name: "react-dom",
      global: "ReactDOM",
      primary:   "https://unpkg.com/react-dom@18.3.1/umd/react-dom.production.min.js",
      fallback:  "https://cdn.jsdelivr.net/npm/react-dom@18.3.1/umd/react-dom.production.min.js",
      integrity: "sha384-gTGxhz21lVGYNMcdJOyq01Edg0jhn/c22nsx0kyqP0TxaV5WVdsSH1fSDUf5YJj1",  // SRI sha384，与 docs/js/resource-manifest.json 保持一致
    },
    {
      kind: "js",
      name: "dayjs",
      global: "dayjs",
      primary:   "https://unpkg.com/dayjs@1.11.13/dayjs.min.js",
      fallback:  "https://cdn.jsdelivr.net/npm/dayjs@1.11.13/dayjs.min.js",
      integrity: "sha384-DpVxUeeBWjUvUV1czyIHJAjh+jYUZFu2lLakbdua5vbwOrBGi1UgaKCHjTC+x3Ky",  // SRI sha384，与 docs/js/resource-manifest.json 保持一致
    },
    {
      kind: "js",
      name: "dayjs locale zh-cn",
      // // no global registration required (side-effect only),
      primary:   "https://unpkg.com/dayjs@1.11.13/locale/zh-cn.js",
      fallback:  "https://cdn.jsdelivr.net/npm/dayjs@1.11.13/locale/zh-cn.js",
      integrity: "sha384-JlOvD9rwLmKdB9EoDKPPtxT02Btalz7oisrWI0tS0cwjMnTvZB9ww5BpcPAnhn+a",  // SRI sha384，与 docs/js/resource-manifest.json 保持一致
    },
    {
      kind: "js",
      name: "@ant-design/icons",
      global: "icons",
      primary:   "https://unpkg.com/@ant-design/icons@5.6.1/dist/index.umd.min.js",
      fallback:  "https://cdn.jsdelivr.net/npm/@ant-design/icons@5.6.1/dist/index.umd.min.js",
      integrity: "sha384-0EWwEDsIH8NINq/z7KdxJDjeP9rCWPurEK0BZnkfhKH9WQGpC8FPbOY0mDyOS2WU",  // SRI sha384，与 docs/js/resource-manifest.json 保持一致
    },
    {
      kind: "js",
      name: "antd",
      global: "antd",
      primary:   "https://unpkg.com/antd@5.21.4/dist/antd.min.js",
      fallback:  "https://cdn.jsdelivr.net/npm/antd@5.21.4/dist/antd.min.js",
      integrity: "sha384-aHcvD9lceo5OXgnbAcHp2RPWwCT3dsRbgapm0ASvhDyAYspgj1NC5k/NNUEsAT6v",  // SRI sha384，与 docs/js/resource-manifest.json 保持一致
    },
  ];


  function once(fn) {
    let done = false; let result;
    return function () {
      if (done) return result;
      done = true;
      result = fn.apply(this, arguments);
      return result;
    };
  }

  function loaded(key) {
    if (!key) return false;
    return Object.prototype.hasOwnProperty.call(global, key) && global[key] != null;
  }

  function injectCSS(href) {
    if (document.querySelector(`link[rel=stylesheet][href="${href}"]`)) return Promise.resolve();
    return new Promise((resolve, reject) => {
      const link = document.createElement('link');
    link.integrity = asset.integrity;
    link.crossOrigin = 'anonymous';
    link.referrerPolicy = 'no-referrer-when-downgrade';
      link.rel = 'stylesheet';
      link.href = href;
      link.onload = () => resolve();
      link.onerror = () => {
        if (link.parentNode) link.parentNode.removeChild(link);
        reject(new Error('CSS failed: ' + href));
      };
      document.head.appendChild(link);
    });
  }

  function injectJS(src) {
    if (document.querySelector(`script[src="${src}"]`)) return Promise.resolve();
    return new Promise((resolve, reject) => {
      const s = document.createElement('script');
    s.integrity = asset.integrity;
    s.crossOrigin = 'anonymous';
    s.referrerPolicy = 'no-referrer-when-downgrade';
      s.src = src;
      s.async = false;
      s.defer = false;
      s.crossOrigin = 'anonymous';
      s.referrerPolicy = 'no-referrer-when-downgrade';
      s.onload = () => resolve();
      s.onerror = () => {
        if (s.parentNode) s.parentNode.removeChild(s);
        reject(new Error('Script failed: ' + src));
      };
      document.head.appendChild(s);
    });
  }

  function tryPrimaryThenFallback(asset) {
    const doInject = asset.kind === 'css' ? injectCSS : injectJS;
    return doInject(asset.primary).catch(() => doInject(asset.fallback));
  }

  async function load(asset) {
    if (asset.kind === 'js' && asset.global && loaded(asset.global)) {
      return { asset, status: 'already-loaded' };
    }
    await tryPrimaryThenFallback(asset);
    return { asset, status: 'loaded' };
  }

  // 串行加载：dayjs antd / icons 的全局变量依赖 React/ReactDOM 先就绪
  async function boot() {
    for (const asset of ASSETS) await load(asset);
  }

  global.ArgusCDN = {
    VERSIONS,
    ASSETS,
    PRIMARY,
    FALLBACK,
    ready: once(() => boot().then(() => ({ versions: VERSIONS })))
  };
})(typeof window !== 'undefined' ? window : this);
