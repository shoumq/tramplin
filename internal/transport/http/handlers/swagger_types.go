package handlers

import "tramplin/internal/models"

type SuccessResponse struct {
	Status string `json:"status" example:"ok"`
	Data   any    `json:"data"`
}

type ErrorResponse struct {
	Status string `json:"status" example:"error"`
	Error  string `json:"error" example:"bad request"`
}

type ChatConversationResponse struct {
	Status string                  `json:"status" example:"ok"`
	Data   models.ChatConversation `json:"data"`
}

type ChatConversationListResponse struct {
	Status string                    `json:"status" example:"ok"`
	Data   []models.ChatConversation `json:"data"`
}
