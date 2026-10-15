(function () {
  const DEFAULT_URL = "https://todo.strbagus.com";
  // const DEFAULT_URL = "http://localhost:8080";
  const frame = document.getElementById("app-frame");

  function loadFrame(url) {
    if (frame) {
      frame.src = url;
    }
  }

  if (chrome && chrome.storage && chrome.storage.sync) {
    chrome.storage.sync.get({ targetUrl: DEFAULT_URL }, function (data) {
      loadFrame(data.targetUrl || DEFAULT_URL);
    });
  } else {
    loadFrame(DEFAULT_URL);
  }
})();
