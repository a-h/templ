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

  document
    .getElementById("content-container")
    .classList.toggle("overflow-hidden");

  let body = document.getElementsByTagName("body")[0];
  body.classList.toggle("overflow-hidden");
}

document.addEventListener("DOMContentLoaded", () => {
  let dropdowns = document.querySelectorAll('div[id^="section-dropdown-"]');
  if (dropdowns.length === 0) {
    return;
  }

  dropdowns.forEach((d) => {
    d.addEventListener("click", handleDropdownClick);

    let children = getChildrenRecursively(d);

    children.forEach((c) => {
      c.addEventListener("click", handleDropdownClick);
    });
  });
});

function handleDropdownClick(event) {
  // stopPropagation() to prevent
  // <path> elems within <svg> from triggering twice
  event.stopPropagation();
  let navItem = event.target;

  while (!navItem.id.startsWith("section-container")) {
    navItem = navItem.parentElement;
  }

  navItem.classList.toggle("active");

  let others = document.querySelectorAll('div[id^="section-container-"]');
  others.forEach((other) => {
    if (other !== navItem) {
      other.classList.remove("active");
    }
  });
}

function getChildrenRecursively(element, descendants = []) {
  element.childNodes.forEach((child) => {
    descendants.push(child);
    getChildrenRecursively(child, descendants);
  });
  return descendants;
}
