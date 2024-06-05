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
document.addEventListener("htmx:load", () => {
  const toggles = document.querySelectorAll(".toggle");
  if (toggles.length !== 0) {
    toggles.forEach((toggle) => {
      toggle.removeEventListener("click", handleToggle);
      toggle.addEventListener("click", handleToggle);
    });
  }
});
