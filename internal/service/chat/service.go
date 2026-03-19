package chat

import (
	"strings"

	"tramplin/internal/dto"
	"tramplin/internal/models"
	"tramplin/internal/repository"
)

type Service struct{ repo repository.PlatformRepository }

func New(repo repository.PlatformRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateConversation(userID string, input dto.ChatConversationInput) (*models.ChatConversation, error) {
	return s.repo.CreateChatConversation(userID, input.ParticipantUserID, input.OpportunityID)
}

func (s *Service) ListConversations(userID string) ([]models.ChatConversation, error) {
	return s.repo.ListChatConversations(userID)
}

func (s *Service) GetConversation(userID, conversationID string) (*models.ChatConversation, error) {
	return s.repo.GetChatConversation(userID, conversationID)
}

func (s *Service) ListMessages(userID, conversationID string) ([]models.ChatMessage, error) {
	return s.repo.ListChatMessages(userID, conversationID)
}

func (s *Service) CreateMessage(userID, conversationID string, input dto.ChatMessageInput) (*models.ChatMessage, error) {
	return s.repo.CreateChatMessage(userID, conversationID, strings.TrimSpace(input.Body))
}

func (s *Service) MarkMessagesRead(userID, conversationID string) (int64, error) {
	return s.repo.MarkChatMessagesRead(userID, conversationID)
}

func (s *Service) TouchPresence(userID string, isOnline bool) error {
	return s.repo.TouchUserPresence(userID, isOnline)
}

func (s *Service) GetUserPresence(userID string) (*models.Presence, error) {
	return s.repo.GetUserPresence(userID)
}

func (s *Service) GetCompanyPresence(companyID string) (*models.Presence, error) {
	return s.repo.GetCompanyPresence(companyID)
}
