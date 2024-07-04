function darkModeToggle() {
  let theme = localStorage.theme;
  if (theme === "dark") {
    localStorage.theme = "light";
    document.documentElement.classList.remove("dark");
  } else {
    localStorage.theme = "dark";
    document.documentElement.classList.add("dark");
  }
}

function toggleSidebar() {
  document.getElementById("sidebar-container").classList.toggle("hidden");
  document.getElementById("sidebar-container").classList.toggle("flex");

  document.getElementById("overlay").classList.toggle("hidden");
  document.getElementById("overlay").classList.toggle("opacity-70");
  
  document.getElementById("content-container").classList.toggle("overflow-hidden")

  let body = document.getElementsByTagName("body")[0];
  body.classList.toggle("overflow-hidden");
}

document.addEventListener("htmx:load", () => {
  const toggles = document.querySelectorAll(".toggle");
  if (toggles.length !== 0) {
    toggles.forEach((toggle) => {
      toggle.removeEventListener("click", handleToggle);
      toggle.addEventListener("click", handleToggle);
    });
  }
});

function handleToggle ({ target }) {
  let navItem = target.parentElement;
  navItem.classList.toggle("active");

  let others = document.querySelectorAll(".toggle");
  others.forEach((other) => {
    if (other !== target) {
      other.parentElement.classList.remove("active");
    }
  });

  // no default action
  return false;
};

// prism-plugin-templ.js runs but it doesn't update
// the dom until after a manual refresh. This makes gives
// the templ language its colors without needing to refresh.
document.addEventListener("DOMContentLoaded", () => {
  Prism.highlightAll();  
})

document.addEventListener("htmx:load", () => {
  Prism.highlightAll();  
})
