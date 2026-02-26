package apierrors

import (
	"github.com/lambdadb/go-lambdadb/models/components"
)

type BadRequestError struct {
	Message  *string                 `json:"message,omitzero"`
	HTTPMeta components.HTTPMetadata `json:"-"`
}

var _ error = &BadRequestError{}

func (e *BadRequestError) Error() string {
	return formatErrorMessage(e.Message, "bad request")
}
