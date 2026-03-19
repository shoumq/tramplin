package dto

type ChatConversationInput struct {
	ParticipantUserID string `json:"participant_user_id"`
	OpportunityID     string `json:"opportunity_id"`
}

type ChatMessageInput struct {
	Body string `json:"body"`
}
