package apierrors

import (
	"testing"
)

func TestFormatErrorMessage(t *testing.T) {
	tests := []struct {
		msg        *string
		defaultMsg string
		want       string
	}{
		{nil, "default", "default"},
		{strPtr(""), "default", "default"},
		{strPtr("custom"), "default", "custom"},
	}
	for _, tt := range tests {
		got := formatErrorMessage(tt.msg, tt.defaultMsg)
		if got != tt.want {
			t.Errorf("formatErrorMessage(%v, %q) = %q, want %q", tt.msg, tt.defaultMsg, got, tt.want)
		}
	}
}

func strPtr(s string) *string { return &s }

func TestTypedErrors_Error(t *testing.T) {
	customMsg := "invalid collection name"
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"BadRequestError with message", &BadRequestError{Message: &customMsg}, "invalid collection name"},
		{"BadRequestError without message", &BadRequestError{}, "bad request"},
		{"ResourceNotFoundError with message", &ResourceNotFoundError{Message: &customMsg}, "invalid collection name"},
		{"ResourceNotFoundError without message", &ResourceNotFoundError{}, "resource not found"},
		{"UnauthenticatedError without message", &UnauthenticatedError{}, "unauthenticated"},
		{"TooManyRequestsError without message", &TooManyRequestsError{}, "too many requests"},
		{"InternalServerError without message", &InternalServerError{}, "internal server error"},
		{"ResourceAlreadyExistsError without message", &ResourceAlreadyExistsError{}, "resource already exists"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}
