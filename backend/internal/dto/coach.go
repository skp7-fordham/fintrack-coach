package dto

// CoachChatRequest is the body for POST /coach/chat.
type CoachChatRequest struct {
	Message        string  `json:"message"`
	ConversationID *string `json:"conversation_id"`
}
