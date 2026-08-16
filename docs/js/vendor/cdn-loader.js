/* 统一维护 Argus Docs 的 CDN 依赖源：
 *   1) 默认 unpkg（Ant Design 文档推荐），失败时走 jsdelivr（与 antd 官网上的『国内镜像』思路一致）
 *   2) 仅在 window 上没有对应全局变量时才追加脚本，避免重复加载
 *   3) 所有资源版本号 pin 死在一个对象上，方便后续升级一处改
 */
(function (global) {
  'use strict';

  var VERSIONS = {
    react: '18.3.1',
    antd: '5.21.4',
    icons: '5.6.1',
    dayjs: '1.11.13'
  };

  var PRIMARY = 'https://unpkg.com';
  var FALLBACK = 'https://cdn.jsdelivr.net/npm';

  var ASSETS = [
    {
      kind: 'css',
      name: 'antd reset.css',
      primary: 'https://unpkg.com/antd@5.21.4/dist/reset.css',
      fallback: 'https://cdn.jsdelivr.net/npm/antd@5.21.4/dist/reset.css'
    },
    {
      kind: 'js',
      name: 'react',
      global: 'React',
      primary: 'https://unpkg.com/react@18.3.1/umd/react.production.min.js',
      fallback: 'https://cdn.jsdelivr.net/npm/react@18.3.1/umd/react.production.min.js'
    },
    {
      kind: 'js',
      name: 'react-dom',
      global: 'ReactDOM',
      primary: 'https://unpkg.com/react-dom@18.3.1/umd/react-dom.production.min.js',
      fallback: 'https://cdn.jsdelivr.net/npm/react-dom@18.3.1/umd/react-dom.production.min.js'
    },
    {
      kind: 'js',
      name: 'dayjs',
      global: 'dayjs',
      primary: 'https://unpkg.com/dayjs@1.11.13/dayjs.min.js',
      fallback: 'https://cdn.jsdelivr.net/npm/dayjs@1.11.13/dayjs.min.js'
    },
    {
      kind: 'js',
      name: 'dayjs locale zh-cn',
      primary: 'https://unpkg.com/dayjs@1.11.13/locale/zh-cn.js',
      fallback: 'https://cdn.jsdelivr.net/npm/dayjs@1.11.13/locale/zh-cn.js'
    },
    {
      kind: 'js',
      name: '@ant-design/icons',
      global: 'icons',
      primary: 'https://unpkg.com/@ant-design/icons@5.6.1/dist/index.umd.min.js',
      fallback: 'https://cdn.jsdelivr.net/npm/@ant-design/icons@5.6.1/dist/index.umd.min.js'
    },
    {
      kind: 'js',
      name: 'antd',
      global: 'antd',
      primary: 'https://unpkg.com/antd@5.21.4/dist/antd.min.js',
      fallback: 'https://cdn.jsdelivr.net/npm/antd@5.21.4/dist/antd.min.js'
    }
  ];

  function once(fn) {
    var done = false;
    var result;
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

  /* injectCSS: 注入 CSS link 标签到 head
   * asset 参数从外层闭包传入，如果带 integrity 就加 SRI 校验
   */
  function injectCSS(href, asset) {
    if (document.querySelector('link[rel="stylesheet"][href="' + href + '"]')) return Promise.resolve();
    return new Promise(function (resolve, reject) {
      var link = document.createElement('link');
      if (asset && asset.integrity) {
        link.integrity = asset.integrity;
        link.crossOrigin = 'anonymous';
      }
      link.referrerPolicy = 'no-referrer-when-downgrade';
      link.rel = 'stylesheet';
      link.href = href;
      link.onload = function () { resolve(); };
      link.onerror = function () {
        if (link.parentNode) link.parentNode.removeChild(link);
        reject(new Error('CSS failed: ' + href));
      };
      document.head.appendChild(link);
    });
  }

  /* injectJS: 注入 script 标签到 head
   * asset 参数从外层闭包传入，如果带 integrity 就加 SRI 校验
   */
  function injectJS(src, asset) {
    if (document.querySelector('script[src="' + src + '"]')) return Promise.resolve();
    return new Promise(function (resolve, reject) {
      var s = document.createElement('script');
      if (asset && asset.integrity) {
        s.integrity = asset.integrity;
        s.crossOrigin = 'anonymous';
      }
      s.referrerPolicy = 'no-referrer-when-downgrade';
      s.src = src;
      s.async = false;
      s.defer = false;
      s.onload = function () { resolve(); };
      s.onerror = function () {
        if (s.parentNode) s.parentNode.removeChild(s);
        reject(new Error('Script failed: ' + src));
      };
      document.head.appendChild(s);
    });
  }

  /* tryPrimaryThenFallback: 先试 primary CDN，失败后自动回落到 fallback CDN
   * 把 asset 传给 injectCSS/injectJS，让 integrity（如果有）生效
   */
  function tryPrimaryThenFallback(asset) {
    var doInject = asset.kind === 'css' ? injectCSS : injectJS;
    return doInject(asset.primary, asset).catch(function () {
      return doInject(asset.fallback, asset);
    });
  }

  /* load: 加载单个资源，如果全局变量已存在则跳过
   */
  function load(asset) {
    if (asset.kind === 'js' && asset.global && loaded(asset.global)) {
      return Promise.resolve({ asset: asset, status: 'already-loaded' });
    }
    return tryPrimaryThenFallback(asset).then(function () {
      return { asset: asset, status: 'loaded' };
    });
  }

  /* boot: 串行加载所有资源（React → ReactDOM → dayjs → icons → antd）
   * 串行是因为 antd 依赖 React/ReactDOM 先就绪
   */
  function boot() {
    var chain = Promise.resolve();
    ASSETS.forEach(function (asset) {
      chain = chain.then(function () { return load(asset); });
    });
    return chain;
  }

  global.ArgusCDN = {
    VERSIONS: VERSIONS,
    ASSETS: ASSETS,
    PRIMARY: PRIMARY,
    FALLBACK: FALLBACK,
    ready: once(function () {
      return boot().then(function () { return { versions: VERSIONS }; });
    })
  };
})(typeof window !== 'undefined' ? window : this);
