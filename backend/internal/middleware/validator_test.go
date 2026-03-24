package middleware

import (
	"testing"
)

type testPayload struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
}

func TestCustomValidator_Validate(t *testing.T) {
	cv := NewValidator()

	t.Run("valid struct passes validation", func(t *testing.T) {
		payload := testPayload{
			Name:  "Alice",
			Email: "alice@example.com",
		}

		if err := cv.Validate(payload); err != nil {
			t.Errorf("expected no error for valid struct, got %v", err)
		}
	})

	t.Run("invalid struct returns error", func(t *testing.T) {
		payload := testPayload{
			Name:  "",
			Email: "not-an-email",
		}

		err := cv.Validate(payload)
		if err == nil {
			t.Fatal("expected error for invalid struct, got nil")
		}
	})
}

func TestNewValidator(t *testing.T) {
	t.Run("creates a working validator", func(t *testing.T) {
		cv := NewValidator()
		if cv == nil {
			t.Fatal("expected non-nil validator")
		}

		// Verify it works by validating a valid struct
		valid := testPayload{Name: "Bob", Email: "bob@test.com"}
		if err := cv.Validate(valid); err != nil {
			t.Errorf("new validator should validate correctly, got %v", err)
		}

		// And catches invalid input
		invalid := testPayload{}
		if err := cv.Validate(invalid); err == nil {
			t.Error("new validator should reject invalid struct")
		}
	})
}
