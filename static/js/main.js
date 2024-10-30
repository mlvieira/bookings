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

const drawDatePicker = (elem = document, removeDisable = true) => {
	const roomsPage = elem.querySelector('#availability-form-container');
	if (roomsPage) return;

	const formBooking = elem.querySelector('#reservation-dates');
	if (!formBooking) return;

	const datePicker = new DateRangePicker(formBooking, {
		format: 'mm-dd-yyyy',
		todayHighlight: true,
		clearButton: true,
		buttonClass: 'btn',
		container: formBooking

	});

	if (removeDisable) {
		removeDisabledDate(elem);
	}

	return datePicker
};

const removeDisabledDate = (elem = document) => {
	elem.querySelectorAll('.form-control')?.forEach(button => { button.removeAttribute('disabled') });
}

const Prompt = () => {
	const showAlert = (config) => {
		const {
			msg = "",
			title = "",
			footer = "",
			icon = "info",
			position = "center",
			timer = null,
			showConfirmButton = true
		} = config;

		Swal.fire({
			icon,
			title,
			text: msg,
			footer,
			position,
			showConfirmButton,
			timer,
			timerProgressBar: !!timer
		});
	};

	const toast = async ({ msg = "", icon = "success", position = "top-end" }) => {
		const Toast = Swal.mixin({
			toast: true,
			position,
			showConfirmButton: false,
			timer: 3000,
			customClass: {
				popup: 'bg-white border-0 shadow-sm position-relative'
			},
			timerProgressBar: true,
			didOpen: (toast) => {
				const closeButton = document.createElement('button');
				closeButton.classList.add('btn-close', 'position-absolute', 'top-0', 'end-0', 'm-2');
				closeButton.addEventListener('click', () => Swal.close());
				toast.appendChild(closeButton);

				toast.addEventListener('mouseenter', Swal.stopTimer);
				toast.addEventListener('mouseleave', Swal.resumeTimer);
			}
		});
		await Toast.fire({ icon, title: msg });
	};

	const custom = async (config) => {
		const {
			icon = "info",
			msg = "",
			title = "",
			showConfirmButton = true,
			showCancelButton = false,
			allowOutsideClick = false,
			showLoaderOnConfirm = false,
			willOpen,
			didOpen,
			callback
		} = config;

		const { value: result } = await Swal.fire({
			icon,
			title,
			html: msg,
			focusConfirm: false,
			showCancelButton,
			showConfirmButton,
			allowOutsideClick,
			showLoaderOnConfirm,
			willOpen: willOpen || undefined,
			didOpen: didOpen || undefined
		});

		if (callback) {
			if (result && result.dismiss !== Swal.DismissReason.cancel && result.value !== "") {
				callback(result);
			} else {
				callback(false);
			}
		}
	}

	const success = (config) => showAlert({ ...config, icon: "success" });
	const error = (config) => showAlert({ ...config, icon: "error" });

	return {
		success,
		toast,
		error,
		custom
	};
};

const roomAvailability = () => {
	const btn = document.getElementById('search-availability');
	if (!btn) return;

	const alert = Prompt();

	btn.addEventListener('click', (e) => {
		e.preventDefault();
		const formContainer = document.getElementById('availability-form-container');
		const html = formContainer.innerHTML;
		
		alert.custom({
			title: 'Search Availability',
			msg: html,
			allowOutsideClick: true,
			showLoaderOnConfirm: true,
			showConfirmButton: false,
			willOpen: () => {
				drawDatePicker(Swal.getPopup(), false);
			},
			didOpen: () => {
				const modalContent = Swal.getPopup();
				removeDisabledDate(modalContent);

				const form = modalContent.querySelector('#availability-form');
				form.addEventListener('submit', async (e) => {
					e.preventDefault();
					const formData = new FormData(form);

					try {
						const response = await fetch('/availability/json', {
							method: 'POST',
							headers: {
								'X-Requested-With': 'XMLHttpRequest',
							},
							body: formData,
						});

						if (!response.ok) {
							throw new Error('Http error!');
						}

						const data = await response.json();

						if (data.ok) {
							alert.success({ msg: data.message });
						} else {
							alert.error({ msg: data.message });
						}
					} catch (error) {
						alert.error({ msg: 'An error occurred while checking availability.' });
						console.error('Error: ', error);
					}
				});
			}
		});
	});
};

const displayMessages = async () => {
	const popUp = document.querySelectorAll('.do-popup');
	if (!popUp) return;

	const alert = Prompt();

	for (const err of popUp) {
		await alert.toast({
			msg: err.dataset.message,
			icon: err.dataset.class,
		});
	};
};

document.addEventListener('DOMContentLoaded', () => {
	'use strict'

	drawDatePicker();
	roomAvailability();
	displayMessages();
});