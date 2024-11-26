document.addEventListener('DOMContentLoaded', () => {
    const dataTable = new simpleDatatables.DataTable("#tableadmin", {
	    searchable: true,
	    fixedHeight: true,
        columns: [
            {
                select: 3,
                sort: "desc",
                format: "MM-DD-YYYY"
            }
        ],
    });
});