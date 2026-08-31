// Light/dark theme toggle for Studigame.
// Self-served static file (Content-Security-Policy script-src 'self').
// Applies [data-theme="dark"] on <html>; persists the choice in localStorage
// and falls back to the system color scheme.
(() => {
  var KEY = "studigame-theme";

  function apply(theme) {
    if (theme === "dark") document.documentElement.setAttribute("data-theme", "dark");
    else document.documentElement.removeAttribute("data-theme");
    var toggle = document.querySelector("[data-theme-toggle]");
    if (toggle) toggle.setAttribute("aria-pressed", theme === "dark" ? "true" : "false");
  }

  function prefersDark() {
    return window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches;
  }

  // Initial paint: stored choice wins, else system preference.
  apply(localStorage.getItem(KEY) || (prefersDark() ? "dark" : "light"));

  document.addEventListener("click", (e) => {
    var btn = e.target.closest("[data-theme-toggle]");
    if (!btn) return;
    var dark = document.documentElement.getAttribute("data-theme") === "dark";
    var next = dark ? "light" : "dark";
    localStorage.setItem(KEY, next);
    apply(next);
  });
})();
