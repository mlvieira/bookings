package forms

import (
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T) {
	postData := url.Values{}
	form := New(postData)

	isValid := form.Valid()
	if !isValid {
		t.Error("got invalid when it should be valid")
	}
}

func TestForm_Required(t *testing.T) {
	postData := url.Values{}
	form := New(postData)

	form.Required("a", "b", "c")
	if form.Valid() {
		t.Error("form shows valid when required fields are missing")
	}

	postData = url.Values{}
	postData.Add("a", "a")
	postData.Add("b", "b")
	postData.Add("c", "c")

	form = New(postData)
	form.Required("a", "b", "c")

	if !form.Valid() {
		t.Error("does not have the required fields")
	}
}

func TestNew(t *testing.T) {
	postData := url.Values{}
	newForm := New(postData)

	if newForm == nil {
		t.Error("expected non-nil pointer, got nil")
	}
}

func TestForm_Has(t *testing.T) {
	postData := url.Values{}
	newForm := New(postData)

	has := newForm.Has("whatever")
	if has {
		t.Error("Form shows it has a field when it does not")
	}

	postData = url.Values{}
	postData.Add("a", "aaa")
	newForm = New(postData)

	has = newForm.Has("a")
	if !has {
		t.Error("Form shows it does not have a field when it does")
	}
}

func TestForm_MinLength(t *testing.T) {
	postData := url.Values{}
	newForm := New(postData)

	newForm.MinLength("non", 10)
	if newForm.Valid() {
		t.Error("Form shows min length for a non-existing field")
	}

	postData = url.Values{}
	postData.Add("test", "1234")
	newForm = New(postData)

	newForm.MinLength("test", 3)

	if !newForm.Valid() {
		t.Error("Expected MinLength validation to pass")
	}

	postData = url.Values{}
	postData.Add("test", "1")
	newForm = New(postData)

	newForm.MinLength("test", 5)
	if newForm.Valid() {
		t.Error("Expected MinLength validation to fail for length < 5")
	}
}

func TestForm_IsEmail(t *testing.T) {
	postData := url.Values{}
	postData.Add("email", "john@example")

	newForm := New(postData)
	newForm.IsEmail("email")

	if newForm.Valid() {
		t.Error("Expected IsEmail validation to fail for an invalid email")
	}

	err := newForm.Errors.Get("email")
	if err == "" {
		t.Error("Expected to return an error for an invalid email")
	}

	postData = url.Values{}
	postData.Add("email", "john@example.com")

	newForm = New(postData)
	newForm.IsEmail("email")

	if !newForm.Valid() {
		t.Error("Expected IsEmail validation to pass for a valid email")
	}

	err = newForm.Errors.Get("email")
	if err != "" {
		t.Error("Expected to not return an error for a valid email")
	}

}
