/* Argus Docs 纯 JS 交互（无框架）
 * 功能：
 *   1. 点击 data-copy 元素复制命令到剪贴板 + toast 提示
 *   2. 平滑滚动锚点
 *   3. KPI 卡片 hover glow 延迟优化
 */
(function () {
  'use strict';

  /* Toast 通知：显示 2 秒后自动隐藏 */
  var toastTimer = null;
  function showToast(msg) {
    var el = document.getElementById('toast');
    if (!el) return;
    el.textContent = msg;
    el.classList.add('show');
    if (toastTimer) clearTimeout(toastTimer);
    toastTimer = setTimeout(function () {
      el.classList.remove('show');
    }, 2000);
  }

  /* 复制到剪贴板：优先 navigator.clipboard，降级 execCommand */
  function copyToClipboard(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(function () {
        showToast('已复制: ' + text);
      }).catch(function () {
        fallbackCopy(text);
      });
    } else {
      fallbackCopy(text);
    }
  }

  function fallbackCopy(text) {
    var ta = document.createElement('textarea');
    ta.value = text;
    ta.style.position = 'fixed';
    ta.style.opacity = '0';
    document.body.appendChild(ta);
    ta.select();
    try {
      document.execCommand('copy');
      showToast('已复制: ' + text);
    } catch (_) {
      showToast('复制失败，请手动复制');
    }
    document.body.removeChild(ta);
  }

  /* 绑定所有 data-copy 元素的点击事件 */
  function bindCopyElements() {
    var els = document.querySelectorAll('[data-copy]');
    els.forEach(function (el) {
      el.addEventListener('click', function () {
        var text = el.getAttribute('data-copy');
        if (text) copyToClipboard(text);
      });
      if (!el.hasAttribute('title')) {
        el.setAttribute('title', '点击复制');
      }
    });
  }

  /* 平滑滚动到锚点 */
  function bindSmoothScroll() {
    var links = document.querySelectorAll('a[href^="#"]');
    links.forEach(function (link) {
      link.addEventListener('click', function (e) {
        var href = link.getAttribute('href');
        if (href === '#' || href.length < 2) return;
        var target = document.querySelector(href);
        if (target) {
          e.preventDefault();
          target.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
      });
    });
  }

  /* DOM Ready */
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function () {
      bindCopyElements();
      bindSmoothScroll();
    });
  } else {
    bindCopyElements();
    bindSmoothScroll();
  }
})();
