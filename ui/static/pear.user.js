// ==UserScript==
// @name         Pear
// @namespace    https://pear.dunkirk.sh
// @version      1.5
// @description  Detect recipe pages and offer to clean them with Pear
// @author       you
// @match        *://*/*
// @grant        none
// @run-at       document-idle
// @downloadURL  https://pear.dunkirk.sh/static/pear.user.js
// @updateURL    https://pear.dunkirk.sh/static/pear.user.js
// ==/UserScript==

(function () {
const PEAR_URL = "https://pear.dunkirk.sh";

  function findRecipe(v) {
    if (typeof v !== "object" || v === null) return false;
    if (Array.isArray(v)) return v.some(findRecipe);
    const types = [v["@type"] ?? []].flat();
    if (types.some((t) => (typeof t === "string" ? t : "").includes("Recipe"))) return true;
    if (v["@graph"] && findRecipe(v["@graph"])) return true;
    for (const val of Object.values(v)) {
      if (typeof val === "object" && val !== null && findRecipe(val)) return true;
    }
    return false;
  }

  function hasRecipeLD() {
    for (const el of document.querySelectorAll('script[type="application/ld+json"]')) {
      try {
        if (findRecipe(JSON.parse(el.textContent))) return true;
      } catch {}
    }
    return false;
  }

  if (!hasRecipeLD()) return;

  const BAR_H = 50;
  const bar = document.createElement("div");
  bar.id = "pear-bar";
  bar.innerHTML = `
    <span id="pear-label">I found a recipe! Do you wish to</span>
    <a href="${PEAR_URL}/?url=${encodeURIComponent(window.location.href)}" id="pear-link">open it in Pear?</a>
    <button id="pear-close" title="Dismiss">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><path d="M18 6L6 18M6 6l12 12"/></svg>
    </button>
  `;

  const style = document.createElement("style");
  style.textContent = `
    html { margin-top: ${BAR_H}px !important; }
    #pear-bar {
      position: fixed;
      top: 0; left: 0; right: 0;
      z-index: 2147483647;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
      height: ${BAR_H}px;
      padding: 0 40px 0 16px;
      background: #1a1a2e;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
      box-shadow: 0 2px 12px rgba(0,0,0,0.2);
      box-sizing: border-box;
    }
    #pear-label {
      color: rgba(255,255,255,0.5);
      font-size: 13px;
      font-weight: 400;
    }
    #pear-link {
      color: #e85d04;
      font-size: 13px;
      font-weight: 600;
      text-decoration: none;
      transition: color 0.15s;
    }
    #pear-link:hover { color: #ff7b2e; text-decoration: underline; }
    #pear-close {
      position: absolute;
      right: 10px;
      top: 50%;
      transform: translateY(-50%);
      background: none;
      border: none;
      color: rgba(255,255,255,0.3);
      cursor: pointer;
      padding: 6px;
      border-radius: 4px;
      display: flex;
      align-items: center;
      justify-content: center;
      transition: color 0.15s, background 0.15s;
    }
    #pear-close:hover { color: rgba(255,255,255,0.7); background: rgba(255,255,255,0.08); }
  `;
  document.head.appendChild(style);
  document.body.appendChild(bar);

  const shifted = [];
  document.querySelectorAll("*").forEach((el) => {
    if (el.id === "pear-bar" || el.closest("#pear-bar")) return;
    const cs = getComputedStyle(el);
    if (cs.position === "fixed" || cs.position === "sticky") {
      const orig = el.getAttribute("style") || "";
      const current = parseFloat(cs.top) || 0;
      el.style.setProperty("top", (current + BAR_H) + "px", "important");
      shifted.push({ el, orig });
    }
  });

  document.getElementById("pear-close").addEventListener("click", () => {
    bar.remove();
    style.remove();
    shifted.forEach(({ el, orig }) => {
      if (orig) { el.setAttribute("style", orig); } else { el.removeAttribute("style"); }
    });
  });
})();