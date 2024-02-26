document.addEventListener('DOMContentLoaded', () => {
    let timeout = null;

    const search = document.getElementById('search');
    if (search !== null) {
        search.addEventListener('keyup', ({target}) => {
            clearTimeout(timeout);

            let menu = document.getElementById('sidebar');
            menu.classList.remove('open');

            timeout = setTimeout(() => {
                let value = target.value.toLowerCase();

                if (value === '') {
                    document.getElementById("main").classList.remove('hidden');
                    document.getElementById("search-results").classList.add('hidden');
                    document.getElementById("search-clear").classList.add('hidden');
                    return;
                }

                let results = [];
                let excerptWidth = 100;

                for (let i = 0; i < index.length; i++) {
                    let pos = index[i].body.toLowerCase().indexOf(value);
                    if (pos !== -1) {
                        let excerptPrefix = "";
                        if (pos - excerptWidth < 0) {
                            excerptPrefix = index[i].body.substring(0, pos);
                        } else {
                            excerptPrefix = "..." + index[i].body.substring(pos - excerptWidth, pos);
                        }

                        let excerptSuffix = "";
                        if (pos + excerptWidth > index[i].body.length) {
                            excerptSuffix = index[i].body.substring(pos, index[i].body.length);
                        } else {
                            excerptSuffix = index[i].body.substring(pos + value.length, pos + value.length + excerptWidth) + "...";
                        }

                        let term = index[i].body.substring(pos, pos + value.length);

                        results.push(searchresult(index[i].title, index[i].href, excerptPrefix, term, excerptSuffix, pos));
                    }
                }

                if (results.length === 0) {
                    document.getElementById("search-results-list").innerHTML = '';
                    let noResults = document.createElement('div');
                    noResults.classList.add('py-8', 'font-semibold', 'leading-6', 'text-gray-900');
                    let text = document.createTextNode("No results found for '" + value + "'");
                    noResults.appendChild(text);

                    document.getElementById("search-results-list").appendChild(noResults);
                } else {
                    document.getElementById("search-results-list").innerHTML = '';
                    results.forEach(result => {
                        document.getElementById("search-results-list").appendChild(result);
                    });
                }
                document.getElementById("main").classList.add('hidden');
                document.getElementById("search-results").classList.remove('hidden');
                document.getElementById("search-clear").classList.remove('hidden');
            }, 400);
        });
    }

    const searchClear = document.getElementById('search-clear');
    if (searchClear !== null) {
        searchClear.addEventListener('click', () => {
            document.getElementById("main").classList.remove('hidden');
            document.getElementById("search-results").classList.add('hidden');
            document.getElementById("search").value = "";
            document.getElementById("search-clear").classList.add('hidden');
        });
    }
});

function searchresult(title, url, excerptPrefix, term, excerptSuffix, pos) {
    let a = document.createElement('a');
    a.setAttribute('href', base_url + url);

    let div = document.createElement('div');
    div.classList.add('flex', 'flex-col', 'gap-x-6', 'mt-2', 'p-5', 'bg-gray-100', 'rounded', 'hover:bg-gray-200');

    let titleDiv = document.createElement('div');
    titleDiv.classList.add('min-w-0', 'flex-auto');

    let titleP = document.createElement('p');
    titleP.classList.add('text-sm', 'font-semibold', 'leading-6', 'text-gray-900');

    let excerptP = document.createElement('p');
    excerptP.classList.add('mt-1', 'text-xs', 'leading-5', 'text-gray-500');

    let urlDiv = document.createElement('div');
    urlDiv.classList.add('shrink-0', 'sm:flex', 'sm:flex-col', 'sm:items-end');

    let urlP = document.createElement('p');
    urlP.classList.add('mt-1', 'text-xs', 'leading-5', 'text-gray-500');

    let text = document.createTextNode(title);
    titleP.appendChild(text);

    text = document.createTextNode(excerptPrefix);
    excerptP.appendChild(text);

    let strong = document.createElement('strong');
    text = document.createTextNode(term);
    strong.appendChild(text);
    excerptP.appendChild(strong);

    text = document.createTextNode(excerptSuffix);
    excerptP.appendChild(text);

    text = document.createTextNode(url);
    urlP.appendChild(text);

    titleDiv.appendChild(titleP);
    div.appendChild(titleDiv);
    div.appendChild(excerptP);
    urlDiv.appendChild(urlP);
    div.appendChild(urlDiv);
    a.appendChild(div);

    return a;
}