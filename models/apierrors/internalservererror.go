package apierrors

import (
	"github.com/lambdadb/go-lambdadb/models/components"
)

type InternalServerError struct {
	Message  *string                 `json:"message,omitzero"`
	HTTPMeta components.HTTPMetadata `json:"-"`
}

var _ error = &InternalServerError{}

func (e *InternalServerError) Error() string {
	return formatErrorMessage(e.Message, "internal server error")
}
