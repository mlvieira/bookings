package forms

// errors represents a map of form fields to a list of error messages
type errors map[string][]string

// Add appends an error message for a given form field
func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// Get retrieves the first error message for a given form field
func (e errors) Get(field string) string {
	es := e[field]

	if len(es) == 0 {
		return ""
	}

	return es[0]
}
