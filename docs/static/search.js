document.addEventListener("htmx:beforeOnLoad", () => {
  const search = document.getElementById("search");
  if (search !== null) {
    search.removeEventListener("keyup", handleSearch);
  }

  const searchClear = document.getElementById("search-clear");
  if (searchClear !== null) {
    searchClear.removeEventListener("click", handleClearSearch);
  }
});
document.addEventListener("htmx:load", () => {
  const search = document.getElementById("search");
  if (search !== null) {
    search.addEventListener("keyup", handleSearch);
  }

  const searchClear = document.getElementById("search-clear");
  if (searchClear !== null) {
    searchClear.addEventListener("click", handleClearSearch);
  }
});

function handleClearSearch() {
  document.getElementById("main").classList.remove("hidden");
  document.getElementById("search-results").classList.add("hidden");
  document.getElementById("search").value = "";
  document.getElementById("search-clear").classList.add("hidden");
}

function handleSearch(evt) {
  clearTimeout(window.searchTimeout);

  let menu = document.getElementById("sidebar");
  menu.classList.remove("open");

  window.searchTimeout = setTimeout(() => {
    let value = evt.target.value.toLowerCase();

    if (value === "") {
      document.getElementById("main").classList.remove("hidden");
      document.getElementById("search-results").classList.add("hidden");
      document.getElementById("search-clear").classList.add("hidden");
      return;
    }

    let results = [];
    let excerptWidth = 100;

    for (let i = 0; i < AllPagesData.length; i++) {
      let pos = AllPagesData[i].body.toLowerCase().indexOf(value);
      if (pos === -1) {
        continue
      }
      let excerptPrefix = "";
      if (pos - excerptWidth < 0) {
        excerptPrefix = AllPagesData[i].body.substring(0, pos);
      } else {
        excerptPrefix =
          "..." + AllPagesData[i].body.substring(pos - excerptWidth, pos);
      }

      let excerptSuffix = "";
      if (pos + excerptWidth > AllPagesData[i].body.length) {
        excerptSuffix = AllPagesData[i].body.substring(pos, AllPagesData[i].body.length);
      } else {
        excerptSuffix =
          AllPagesData[i].body.substring(
            pos + value.length,
            pos + value.length + excerptWidth,
          ) + "...";
      }

      let term = AllPagesData[i].body.substring(pos, pos + value.length);

      results.push(
        resultMarkup(
          AllPagesData[i].title,
          AllPagesData[i].href,
          escapeHtml(excerptPrefix),
          escapeHtml(term),
          escapeHtml(excerptSuffix),
        ),
      );
    }

    let searchResultsList = document.getElementById("search-results-list")
    let r = ""

    if (results.length > 0) {
        r = results.join("\n")
    } else {
      r = `
        <div class="py-8 font-semibold leading-6 text-gray-900 dark:text-gray-100">
                No results found for "${escapeHtml(value)}"
        </div>
      `
    }
    
    searchResultsList.innerHTML = r

    document.getElementById("main").classList.add("hidden");
    document.getElementById("search-results").classList.remove("hidden");
    document.getElementById("search-clear").classList.remove("hidden");
  }, 400);
}

function resultMarkup(title, url, excerptPrefix, term, excerptSuffix) {
  return `
        <a href="/${url}">
          <div class="flex flex-col gap-x-6 mt-2 p-5 bg-gray-100 rounded hover:bg-gray-200 dark:bg-gray-900 dark:hover:bg-gray-600">
            <div class="min-w-0 flex-auto">
              <p class="text-sm font-semibold leading-6 text-gray-900 dark:text-gray-100">${title}</p>
            </div>
            <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-gray-200">${excerptPrefix}<strong>${term}</strong>${excerptSuffix}</p>
            <div class="shrink-0 sm:flex sm:flex-col sm:items-end">
              <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-gray-400">${url}</p>
            </div>
          </div>
        </a>
  `
}

function escapeHtml(str) {
  return str
    // Escape HTML for safety
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    // Remove markdown symbols for looks
    .replace(/#/g, '')
    .replace(/`/g, '');
}
