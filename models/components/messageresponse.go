package components

type MessageResponse struct {
	Message string `json:"message"`
}

func (m *MessageResponse) GetMessage() string {
	if m == nil {
		return ""
	}
	return m.Message
}
