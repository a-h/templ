document.addEventListener('DOMContentLoaded', () => {
    const toggles = document.querySelectorAll('.toggle');
    if (toggles.length !== 0) {
        toggles.forEach(toggle =>
            toggle.addEventListener('click', ({target}) => {
                let navItem = target.parentElement;
                navItem.classList.toggle('active');

                let others = document.querySelectorAll('.toggle');
                others.forEach(other => {
                    if (other !== toggle) {
                        other.parentElement.classList.remove('active');
                    }
                });

                // no default action
                return false;
            })
        );
    }
});
