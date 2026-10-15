(function () {
  // const DEFAULT_URL = "http://localhost:8080";
  const DEFAULT_URL = "https://todo.strbagus.com";
  const DEFAULT_BG = "/static/img/mywall.jpg";

  const urlInput = document.getElementById("targetUrl");
  const bgInput = document.getElementById("bgUrl");
  const saveBtn = document.getElementById("save-btn");
  const statusEl = document.getElementById("status");

  // Load current
  chrome.storage.sync.get({ targetUrl: DEFAULT_URL, backgroundUrl: DEFAULT_BG }, function (data) {
    urlInput.value = data.targetUrl || DEFAULT_URL;
    bgInput.value = data.backgroundUrl || DEFAULT_BG;
  });

  // Save
  saveBtn.addEventListener("click", function () {
    const targetUrl = urlInput.value.trim() || DEFAULT_URL;
    const backgroundUrl = bgInput.value.trim() || DEFAULT_BG;

    chrome.storage.sync.set({ targetUrl: targetUrl, backgroundUrl: backgroundUrl }, function () {
      statusEl.textContent = "Saved!";
      setTimeout(function () {
        statusEl.textContent = "";
      }, 2000);
    });
  });

  // Preset buttons for target URL
  document.querySelectorAll(".preset-btn").forEach(function (btn) {
    btn.addEventListener("click", function () {
      urlInput.value = btn.getAttribute("data-url");
    });
  });

  // Preset buttons for wallpaper
  document.querySelectorAll(".preset-bg-btn").forEach(function (btn) {
    btn.addEventListener("click", function () {
      bgInput.value = btn.getAttribute("data-bg");
    });
  });
})();
