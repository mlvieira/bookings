document.addEventListener('DOMContentLoaded', () => {
    const calendarEl = document.getElementById('calendar');
    const calendar = new FullCalendar.Calendar(calendarEl, {
        themeSystem: 'bootstrap5',
        navLinks: true,
        editable: false,
        selectable: true,
        dayMaxEvents: true,
        headerToolbar: {
            left: 'prev,next today',
            center: 'title',
            right: 'dayGridMonth,timeGridWeek,timeGridDay,listMonth'
        },
        events: async (fetchInfo, successCallback, failureCallback) => {
            try {
                const response = await fetch('/admin/reservations/calendar/json');
                if (!response.ok) {
                    throw new Error('Failed to fetch events');
                }
                const data = await response.json();
                successCallback(data);
            } catch (error) {
                failureCallback(error);
            }
        },
        eventClick: (info) => {
            const event = info.event;

            const start = event.start ? new Date(event.start).toLocaleString() : "N/A";
            const end = event.end ? new Date(event.end).toLocaleString() : "N/A";
            const lastUpdated = event.extendedProps.lastUpdated
                ? new Date(event.extendedProps.lastUpdated).toLocaleString()
                : "N/A";
            const url = event.url;
            const name = event.extendedProps.name;
            const roomName = event.extendedProps.room;

            const prompt = Prompt();
            prompt.custom({
                title: 'Reservation Details',
                msg: `
                    <p><strong>Room:</strong> ${roomName}</p>
                    <p><strong>Full Name:</strong> ${name}</p>
                    <p><strong>Start:</strong> ${start}</p>
                    <p><strong>End:</strong> ${end}</p>
                    <p><strong>Last Update:</strong> ${lastUpdated}</p>
                `,
                showConfirmButton: true,
                showCancelButton: true,
                allowOutsideClick: true,
                confirmButtonText: "Edit",
                cancelButtonText: "Close",
                callback: (result) => {
                    if (result) {
                        window.location.href = url;
                    }
                },
            });

            info.jsEvent.preventDefault();
        },
    });
    calendar.render();
});