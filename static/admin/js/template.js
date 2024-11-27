document.addEventListener('DOMContentLoaded', () => {
    const mainContent = document.getElementById('main-content');
    const navLinks = document.querySelectorAll('.nav li a');

    const toggleArrow = (collapse, expand) => {
        const arrow = collapse.closest('.nav-item')?.querySelector('.menu-arrow');
        if (arrow) {
            arrow.classList.toggle('fa-arrow-up', expand);
            arrow.classList.toggle('fa-arrow-down', !expand);
        }
    };

    const saveDropdownState = () => {
        const state = {};
        document.querySelectorAll('.collapse').forEach(collapse => {
            state[collapse.id] = collapse.classList.contains('show');
        });
        localStorage.setItem('dropdownState', JSON.stringify(state));
    };

    const restoreDropdownState = () => {
        const savedState = JSON.parse(localStorage.getItem('dropdownState')) || {};
        document.querySelectorAll('.collapse').forEach(collapse => {
            const isExpanded = savedState[collapse.id];
            collapse.classList.toggle('show', isExpanded);
            collapse.closest('.nav-item')?.classList.toggle('active', isExpanded);
            toggleArrow(collapse, isExpanded);
        });
    };

    const updateActiveNav = (url) => {
        const normalizedUrl = new URL(url, window.location.origin).pathname;
        let activeFound = false;

        navLinks.forEach(link => {
            const linkUrl = new URL(link.href, window.location.origin).pathname;
            const navItem = link.closest('.nav-item');
            const collapseParent = link.closest('.sub-menu')?.closest('.collapse');

            const isActive = linkUrl === normalizedUrl && !activeFound;
            link.classList.toggle('active', isActive);
            navItem?.classList.toggle('active', isActive);

            if (collapseParent) {
                const shouldExpand = isActive || collapseParent.querySelector('.active');
                collapseParent.classList.toggle('show', shouldExpand);
                toggleArrow(collapseParent, shouldExpand);
                collapseParent.closest('.nav-item')?.classList.toggle('active', shouldExpand);
            }

            if (isActive) activeFound = true;
        });
    };

    const loadContent = async (url) => {
        mainContent.innerHTML = '<div class="text-center my-5"><div class="spinner-border" role="status"></div></div>';
        try {
            const response = await fetch(url);
            if (!response.ok) throw new Error('Network error');
            const html = await response.text();
            const parser = new DOMParser();
            const doc = parser.parseFromString(html, 'text/html');
            const newContent = doc.querySelector('#main-content')?.innerHTML;

            if (newContent) {
                mainContent.innerHTML = newContent;
                updateActiveNav(url);
                saveDropdownState();
                window.history.pushState({}, '', url);
            }
        } catch (error) {
            const alert = Prompt();
            await alert.toast({
                msg: "Failed to load content",
                icon: "error"
            });
        }
    };

    navLinks.forEach(link => {
        link.addEventListener('click', event => {
            const url = link.getAttribute('href');
            if (url?.startsWith('/')) {
                event.preventDefault();
                loadContent(url);
            }
        });
    });

    document.querySelectorAll('.collapse').forEach(collapse => {
        collapse.addEventListener('show.bs.collapse', () => {
            toggleArrow(collapse, true);
            saveDropdownState();
        });
        collapse.addEventListener('hide.bs.collapse', () => {
            toggleArrow(collapse, false);
            saveDropdownState();
        });
    });

    window.addEventListener('popstate', () => loadContent(location.pathname));

    restoreDropdownState();
    updateActiveNav(location.pathname);

    document.querySelector('.sidebar')?.addEventListener('show.bs.collapse', event => {
        document.querySelectorAll('.collapse.show').forEach(collapse => {
            if (collapse !== event.target) {
                collapse.classList.remove('show');
                toggleArrow(collapse, false);
            }
        });
    });

    document.querySelector('[data-bs-toggle="minimize"]')?.addEventListener('click', () => {
        document.body.classList.toggle('sidebar-icon-only');
        if (document.body.classList.contains('sidebar-icon-only')) {
            document.querySelectorAll('.sidebar .collapse.show').forEach(collapse => {
                collapse.classList.remove('show');
                toggleArrow(collapse, false);
            });
        }
    });

    const sendRequest = async (url, id, alert) => {
        try {
            const response = await fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Requested-With': 'XMLHttpRequest',
                    'X-CSRF-Token': document.querySelector('input[name="csrf_token"]').value,
                },
                body: JSON.stringify({ id }),
            });

            if (!response.ok) {
                throw new Error('Http error!');
            }

            const data = await response.json();
            if (data.ok) {
                alert.success({
                    msg: data.message,
                });
                return true;
            } else {
                alert.error({
                    msg: data.message,
                });
                return false;
            }
        } catch (error) {
            alert.error({
                msg: "An unknown error happened",
            });
            return false;
        }
    };

    const handleAction = (selector, url) => {
        const btn = document.querySelector(selector);
        if (btn) {
            btn.addEventListener('click', (event) => {
                event.preventDefault();
                const alert = Prompt();
                const id = btn.dataset.id;

                alert.custom({
                    msg: "Are you sure you want to confirm?",
                    icon: "warning",
                    allowOutsideClick: true,
                    showCancelButton: true,
                    callback: async () => {
                        const alertResult = await sendRequest(url, id, alert);
                        if (alertResult) {
                            switch (selector) {
                                case '#markProcessed':
                                    btn.disabled = true;
                                    break;
                                case '#deleteRes':
                                    setTimeout(() => {
                                        window.location.href = `/admin/reservations/${btn.dataset.source}`
                                    }, 2 * 1000);
                                    break;
                            }
                        }
                    },
                });
            });
        }
    };
    handleAction('#markProcessed', '/admin/reservations/processed');
    handleAction('#deleteRes', '/admin/reservations/delete');
});
