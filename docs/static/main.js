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

document.addEventListener("DOMContentLoaded", function (_) {
  document.body.addEventListener("htmx:configRequest", function (event) {
    relativeUrl = event.detail.path.split(base_url);

    bodyOnlyUrl = base_url + "body-only/" + relativeUrl[1];

    event.detail.path = bodyOnlyUrl;
    htmx.ajax(event)
  });

  document.addEventListener("htmx:load", (_) => {
    htmx.logAll();
  });
});
