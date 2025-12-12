package server

import "testing"

func TestValidateUserInput_Valid(t *testing.T) {
	errs := validateUserInput("Alice", "alice@example.com", 20)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d", len(errs))
	}
}

func TestValidateUserInput_EmptyName(t *testing.T) {
	errs := validateUserInput("", "alice@example.com", 20)
	if len(errs) == 0 {
		t.Fatalf("expected errors, got none")
	}

	found := false
	for _, e := range errs {
		if e.Field == "name" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error on field 'name'")
	}
}

func TestValidateUserInput_TooLongName(t *testing.T) {
	longName := make([]byte, 260)
	for i := range longName {
		longName[i] = 'a'
	}

	errs := validateUserInput(string(longName), "alice@example.com", 20)
	if len(errs) == 0 {
		t.Fatalf("expected errors, got none")
	}
}

func TestValidateUserInput_EmptyEmail(t *testing.T) {
	errs := validateUserInput("Alice", "", 20)
	if len(errs) == 0 {
		t.Fatalf("expected errors, got none")
	}

	found := false
	for _, e := range errs {
		if e.Field == "email" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error on field 'email'")
	}
}

func TestValidateUserInput_AgeTooSmall(t *testing.T) {
	errs := validateUserInput("Alice", "alice@example.com", 15)
	if len(errs) == 0 {
		t.Fatalf("expected errors, got none")
	}

	found := false
	for _, e := range errs {
		if e.Field == "age" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error on field 'age'")
	}
}
