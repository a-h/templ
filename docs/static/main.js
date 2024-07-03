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
  console.log("toggleSidebar()")
  document.getElementById("sidebar-container").classList.toggle("hidden");
  document.getElementById("sidebar-container").classList.toggle("flex");

  document.getElementById("overlay").classList.toggle("hidden");
  document.getElementById("overlay").classList.toggle("opacity-70");
  
  document.getElementById("content-container").classList.toggle("overflow-hidden")

  let body = document.getElementsByTagName("body")[0];
  body.classList.toggle("overflow-hidden");
}
