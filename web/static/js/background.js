(function () {
  const STORAGE_KEY = "todo_background_url";
  const DEFAULT_BG = "/static/img/mywall.jpg";

  function applyBackground(url) {
    const bg = url || DEFAULT_BG;
    if (bg === "none") {
      document.body.style.backgroundImage = "radial-gradient(ellipse at center, #111726 0%, #060911 100%)";
    } else {
      document.body.style.backgroundImage = `linear-gradient(rgba(7, 10, 18, 0.35), rgba(7, 10, 18, 0.35)), url("${bg}")`;
      document.body.style.backgroundSize = "cover";
      document.body.style.backgroundPosition = "center";
      document.body.style.backgroundRepeat = "no-repeat";
      document.body.style.backgroundAttachment = "fixed";
    }
  }

  // Initial load
  const saved = localStorage.getItem(STORAGE_KEY) || DEFAULT_BG;
  applyBackground(saved);

  // Expose global helpers
  window.setCustomBackground = function (url) {
    localStorage.setItem(STORAGE_KEY, url);
    applyBackground(url);
  };

  window.getCustomBackground = function () {
    return localStorage.getItem(STORAGE_KEY) || DEFAULT_BG;
  };
})();
