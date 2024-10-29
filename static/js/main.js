window.addEventListener('load', () => {
    'use strict'

    const forms = document.querySelectorAll('.needs-validation');

    forms.forEach(form => {
        form.addEventListener('submit', event => {
            if (!form.checkValidity()) {
                event.preventDefault();
                event.stopPropagation();
            }

            form.classList.add('was-validated');
        });
    });
});

const drawDatePicker = () => {
    const formBooking = document.getElementById('reservation-dates');
    if (!formBooking) return;

    const datePicker = new DateRangePicker(formBooking, {
        format: 'mm-dd-yyyy',
        todayHighlight: true,
        clearButton: true,
        buttonClass: 'btn'

    });

    return datePicker
}

document.addEventListener('DOMContentLoaded', () => {
    drawDatePicker();
});