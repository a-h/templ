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
