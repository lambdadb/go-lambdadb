package apierrors

// formatErrorMessage returns the message for API errors. If msg is non-nil and non-empty, it is returned; otherwise defaultMsg is returned.
func formatErrorMessage(msg *string, defaultMsg string) string {
	if msg != nil && *msg != "" {
		return *msg
	}
	return defaultMsg
}
