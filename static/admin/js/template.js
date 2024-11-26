document.addEventListener('DOMContentLoaded', () => {
    'use strict';

    document.querySelectorAll('.nav li a').forEach(link => {
        const currentPage = location.pathname.split('/').pop() || 'index.html';
        if (link.getAttribute('href').includes(currentPage)) {
            link.closest('.nav-item')?.classList.add('active');
            link.closest('.sub-menu')?.closest('.collapse')?.classList.add('show');
            link.classList.add('active');
        }
    });

    document.querySelector('.sidebar')?.addEventListener('show.bs.collapse', (event) => {
        document.querySelectorAll('.collapse.show').forEach(collapse => {
            if (collapse !== event.target) {
                collapse.classList.remove('show');
            }
        });
    });

    document.querySelector('[data-bs-toggle="minimize"]')?.addEventListener('click', () => {
        document.body.classList.toggle('sidebar-icon-only');

        if (document.body.classList.contains('sidebar-icon-only')) {
            document.querySelectorAll('.sidebar .collapse.show').forEach(collapse => {
                collapse.classList.remove('show');
            });
        }
    });
});
