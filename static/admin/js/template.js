document.addEventListener('DOMContentLoaded', () => {
    'use strict';

    const body = document.body;
    const sidebar = document.querySelector('.sidebar');

    const addActiveClass = (element) => {
        const href = element.getAttribute('href');
        if (!href) return;

        const linkPath = new URL(href, window.location.origin).pathname.replace(/\/$/, '');
        const currentPath = window.location.pathname.replace(/\/$/, '') || '/';

        if (linkPath !== currentPath) return;

        element.classList.add('active');

        const collapseElement = element.closest('.collapse');
        if (collapseElement) {
            collapseElement.classList.add('show');

            const parentNavItem = collapseElement.closest('.nav-item');
            if (parentNavItem) {
                parentNavItem.classList.add('active');
            }
        } else {
            const navItem = element.closest('.nav-item');
            if (navItem) {
                navItem.classList.add('active');
            }
        }
    }

    const links = sidebar.querySelectorAll('.nav li a');
    links.forEach(link => {
        addActiveClass(link);
    });

    const collapses = sidebar.querySelectorAll('.collapse');
    collapses.forEach(collapse => {
        collapse.addEventListener('show.bs.collapse', (event) => {
            collapses.forEach(otherCollapse => {
                if (otherCollapse !== event.target && otherCollapse.classList.contains('show')) {
                    const bsCollapse = bootstrap.Collapse.getInstance(otherCollapse);
                    if (bsCollapse) {
                        bsCollapse.hide();
                    }
                }
            });
        });
    });

    const minimizeToggle = document.querySelector('[data-bs-toggle="minimize"]');
    if (minimizeToggle) {
        minimizeToggle.addEventListener('click', () => {
            body.classList.toggle('sidebar-icon-only');
        });
    }

    const sendRequest = async (url, id, alert) => {
        try {
            const response = await fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Requested-With': 'XMLHttpRequest',
                    'X-CSRF-Token': document.querySelector('input[name="csrf_token"]').value,
                },
                body: JSON.stringify({ id: parseInt(id, 10) }),
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
                    callback: async (isConfirmed) => {
                        if (isConfirmed) {
                            const alertResult = await sendRequest(url, id, alert);
                            if (alertResult) {
                                switch (selector) {
                                    case '#markProcessed':
                                        btn.disabled = true;
                                        break;
                                    case '#deleteRes':
                                        setTimeout(() => {
                                            window.location.href = `/admin/reservations/${btn.dataset.source}`;
                                        }, 2000);
                                        break;
                                    case '#deleteUsr':
                                        setTimeout(() => {
                                            window.location.href = `/admin/users`;
                                        }, 2000);
                                        break;
                                }
                            }
                        }
                    },
                });
            });
        }
    };
    handleAction('#markProcessed', '/admin/reservations/processed');
    handleAction('#deleteRes', '/admin/reservations/delete');
    handleAction('#deleteUsr', '/admin/users/delete');
});
