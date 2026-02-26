package apierrors

import (
	"github.com/lambdadb/go-lambdadb/models/components"
)

type TooManyRequestsError struct {
	Message  *string                 `json:"message,omitzero"`
	HTTPMeta components.HTTPMetadata `json:"-"`
}

var _ error = &TooManyRequestsError{}

func (e *TooManyRequestsError) Error() string {
	return formatErrorMessage(e.Message, "too many requests")
}
