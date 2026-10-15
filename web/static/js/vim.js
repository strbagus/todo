// Vim Keybindings for Todo & Deadline Tracker
(function () {
  let selectedIndex = -1;
  let lastKeyTime = 0;
  let lastKey = "";

  function getCards() {
    return Array.from(document.querySelectorAll(".item-card"));
  }

  function updateSelection(newIndex) {
    const cards = getCards();
    if (cards.length === 0) {
      selectedIndex = -1;
      return;
    }

    if (newIndex < 0) newIndex = 0;
    if (newIndex >= cards.length) newIndex = cards.length - 1;

    cards.forEach((card, idx) => {
      if (idx === newIndex) {
        card.classList.add("selected-card");
        card.scrollIntoView({ behavior: "smooth", block: "nearest" });
      } else {
        card.classList.remove("selected-card");
      }
    });

    selectedIndex = newIndex;
  }

  function getSelectedCard() {
    const cards = getCards();
    if (selectedIndex >= 0 && selectedIndex < cards.length) {
      return cards[selectedIndex];
    }
    return null;
  }

  window.addEventListener("keydown", function (e) {
    const activeEl = document.activeElement;
    const isInput =
      activeEl &&
      (activeEl.tagName === "INPUT" ||
        activeEl.tagName === "TEXTAREA" ||
        activeEl.tagName === "SELECT");

    if (e.key === "Escape") {
      if (isInput) {
        activeEl.blur();
        return;
      }
      // Close any visible modal
      const modals = document.querySelectorAll(".modal-backdrop");
      modals.forEach((m) => {
        if (m.id === "edit-modal") {
          m.remove();
        } else {
          m.classList.add("hidden");
        }
      });
      return;
    }

    // Do not intercept keystrokes while typing in form inputs
    if (isInput) {
      return;
    }

    const now = Date.now();
    const isDouble = now - lastKeyTime < 400;

    switch (e.key) {
      case "j":
        e.preventDefault();
        updateSelection(selectedIndex + 1);
        break;

      case "k":
        e.preventDefault();
        updateSelection(selectedIndex - 1);
        break;

      case "g":
        if (isDouble && lastKey === "g") {
          e.preventDefault();
          updateSelection(0);
        }
        break;

      case "G":
        e.preventDefault();
        const cards = getCards();
        updateSelection(cards.length - 1);
        break;

      case "x":
      case " ":
        e.preventDefault();
        const card = getSelectedCard();
        if (card) {
          const toggleBtn = card.querySelector(".item-toggle-btn");
          if (toggleBtn) toggleBtn.click();
        }
        break;

      case "o":
      case "a":
        e.preventDefault();
        if (typeof openAddModal === "function") {
          openAddModal();
        } else {
          const addBtn = document.getElementById("btn-add-item");
          if (addBtn) addBtn.click();
        }
        break;

      case "e":
        e.preventDefault();
        const editCard = getSelectedCard();
        if (editCard) {
          const editBtn = editCard.querySelector(".action-btn.edit");
          if (editBtn) editBtn.click();
        }
        break;

      case "d":
        if (isDouble && lastKey === "d") {
          e.preventDefault();
          const delCard = getSelectedCard();
          if (delCard) {
            const delBtn = delCard.querySelector(".action-btn.delete");
            if (delBtn) delBtn.click();
          }
        }
        break;

      case "/":
        e.preventDefault();
        const searchInput = document.getElementById("search-input");
        if (searchInput) {
          searchInput.focus();
          searchInput.select();
        }
        break;

case "?":
        e.preventDefault();
        const cheatsheet = document.getElementById("cheatsheet-modal");
        if (cheatsheet) {
          cheatsheet.classList.toggle("hidden");
        }
        break;
    }

    lastKey = e.key;
    lastKeyTime = now;
  });

  // Re-sync selection after HTMX content updates
  document.body.addEventListener("htmx:afterSwap", function () {
    const cards = getCards();
    if (selectedIndex >= cards.length) {
      selectedIndex = cards.length - 1;
    }
    if (selectedIndex >= 0) {
      updateSelection(selectedIndex);
    }
  });
})();
