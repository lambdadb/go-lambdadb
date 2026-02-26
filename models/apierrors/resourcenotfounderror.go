package apierrors

import (
	"github.com/lambdadb/go-lambdadb/models/components"
)

type ResourceNotFoundError struct {
	Message  *string                 `json:"message,omitzero"`
	HTTPMeta components.HTTPMetadata `json:"-"`
}

var _ error = &ResourceNotFoundError{}

func (e *ResourceNotFoundError) Error() string {
	return formatErrorMessage(e.Message, "resource not found")
}
