document.addEventListener('DOMContentLoaded', () => {
    const menu_toggles = document.querySelectorAll('.menu-toggle');
    menu_toggles.forEach(menu_toggle => {
        menu_toggle.addEventListener('click', () => {
            toggleSidebar();
        });
    });

    const overlay = document.getElementById('overlay');
    if (overlay !== null) {
        overlay.addEventListener('click', () => {
            if (overlay.classList.contains('opacity-70')) {
                toggleSidebar();
            }
        });
    }

    // On page load or when changing themes, best to add inline in `head` to avoid FOUC
    if (localStorage.theme === 'dark' || (!('theme' in localStorage) && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
        document.documentElement.classList.add('dark')
    } else {
        document.documentElement.classList.remove('dark')
    }

    const dark_mode_toggles = document.querySelectorAll('.dark-mode-toggle');
    dark_mode_toggles.forEach(dark_mode_toggle => {
        dark_mode_toggle.addEventListener('click', () => {
            let theme = localStorage.theme;
            if (theme === 'dark') {
                localStorage.theme = 'light';
            } else {
                localStorage.theme = 'dark';
            }
            document.documentElement.classList.toggle('dark');
        });
    });
});

function toggleSidebar() {
    let menu = document.getElementById('sidebar');
    menu.classList.toggle('open');

    toggleOverlay();

    let body = document.getElementsByTagName('body')[0];
    body.classList.toggle('overflow-hidden');
}

function toggleOverlay() {
    let sidebarOverlay = document.getElementById('overlay');
    sidebarOverlay.classList.toggle('opacity-0');
    sidebarOverlay.classList.toggle('opacity-70');
    setTimeout(() => {
        sidebarOverlay.classList.toggle('hidden');
    }, 200);
}
