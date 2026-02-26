package apierrors

import (
	"github.com/lambdadb/go-lambdadb/models/components"
)

type UnauthenticatedError struct {
	Message  *string                 `json:"message,omitzero"`
	HTTPMeta components.HTTPMetadata `json:"-"`
}

var _ error = &UnauthenticatedError{}

func (e *UnauthenticatedError) Error() string {
	return formatErrorMessage(e.Message, "unauthenticated")
}
