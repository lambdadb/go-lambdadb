package apierrors

import (
	"github.com/lambdadb/go-lambdadb/models/components"
)

type ResourceAlreadyExistsError struct {
	Message  *string                 `json:"message,omitzero"`
	HTTPMeta components.HTTPMetadata `json:"-"`
}

var _ error = &ResourceAlreadyExistsError{}

func (e *ResourceAlreadyExistsError) Error() string {
	return formatErrorMessage(e.Message, "resource already exists")
}
