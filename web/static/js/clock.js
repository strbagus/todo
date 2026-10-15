(function () {
  function updateClock() {
    const now = new Date();
    const timeEl = document.getElementById("clock-time");
    const dateEl = document.getElementById("clock-date");
    const greetingEl = document.getElementById("clock-greeting");

    if (timeEl) {
      const hours = String(now.getHours()).padStart(2, "0");
      const minutes = String(now.getMinutes()).padStart(2, "0");
      timeEl.textContent = `${hours}:${minutes}`;
    }

    if (dateEl) {
      const options = { weekday: "long", month: "long", day: "numeric" };
      dateEl.textContent = now.toLocaleDateString(undefined, options);
    }

    if (greetingEl) {
      const hour = now.getHours();
      let greeting = "Good evening";
      if (hour >= 5 && hour < 12) greeting = "Good morning";
      else if (hour >= 12 && hour < 17) greeting = "Good afternoon";
      greetingEl.textContent = greeting;
    }
  }

  updateClock();
  setInterval(updateClock, 1000);
})();
