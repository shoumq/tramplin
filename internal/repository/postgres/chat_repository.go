package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	. "tramplin/internal/models"
)

func (r *Repository) CreateChatConversation(userID, participantUserID, opportunityID string) (*ChatConversation, error) {
	userID = strings.TrimSpace(userID)
	participantUserID = strings.TrimSpace(participantUserID)
	opportunityID = strings.TrimSpace(opportunityID)
	if userID == "" || participantUserID == "" {
		return nil, errors.New("participant_user_id is required")
	}
	if userID == participantUserID {
		return nil, errors.New("cannot create chat with yourself")
	}
	if _, err := r.getUserByID(context.Background(), participantUserID); err != nil {
		return nil, err
	}

	aID, bID := orderedUserIDs(userID, participantUserID)
	var existingID string
	err := r.db.QueryRowContext(context.Background(), `
SELECT id
FROM chat_conversations
WHERE participant_a_user_id = $1
  AND participant_b_user_id = $2
  AND (
      (opportunity_id = NULLIF($3, '')::uuid)
      OR (opportunity_id IS NULL AND NULLIF($3, '')::uuid IS NULL)
  )
`, aID, bID, opportunityID).Scan(&existingID)
	if err == nil {
		return r.GetChatConversation(userID, existingID)
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("check existing chat conversation: %w", err)
	}

	conversationID := uuid.NewString()
	now := time.Now()
	if _, err := r.db.ExecContext(context.Background(), `
INSERT INTO chat_conversations (
	id, participant_a_user_id, participant_b_user_id, opportunity_id, created_at, updated_at
)
VALUES ($1, $2, $3, NULLIF($4, '')::uuid, $5, $6)
`, conversationID, aID, bID, opportunityID, now, now); err != nil {
		return nil, fmt.Errorf("create chat conversation: %w", err)
	}
	return r.GetChatConversation(userID, conversationID)
}

func (r *Repository) GetChatConversation(userID, conversationID string) (*ChatConversation, error) {
	var item ChatConversation
	var participantName string
	var participantAvatarURL sql.NullString
	var lastMessage sql.NullString
	var lastMessageAt sql.NullTime
	var participantLastSeenAt sql.NullTime
	err := r.db.QueryRowContext(context.Background(), `
SELECT
	c.id,
	COALESCE(c.opportunity_id::text, ''),
	COALESCE(o.title, ''),
	COALESCE(comp.legal_name, ''),
	CASE WHEN c.participant_a_user_id = $1 THEN c.participant_b_user_id ELSE c.participant_a_user_id END AS participant_user_id,
	u.display_name,
	COALESCE(u.avatar_url, ''),
	COALESCE(up.is_online, FALSE),
	up.last_seen_at,
	COALESCE(last_message.body, ''),
	last_message.created_at,
	COALESCE(unread.unread_count, 0),
	c.created_at,
	c.updated_at
FROM chat_conversations c
JOIN users u ON u.id = CASE WHEN c.participant_a_user_id = $1 THEN c.participant_b_user_id ELSE c.participant_a_user_id END
LEFT JOIN user_presence up ON up.user_id = u.id
LEFT JOIN opportunities o ON o.id = c.opportunity_id
LEFT JOIN companies comp ON comp.id = o.company_id
LEFT JOIN LATERAL (
	SELECT body, created_at
	FROM chat_messages
	WHERE conversation_id = c.id
	ORDER BY created_at DESC
	LIMIT 1
) AS last_message ON TRUE
LEFT JOIN LATERAL (
	SELECT COUNT(*) AS unread_count
	FROM chat_messages m
	WHERE m.conversation_id = c.id
	  AND m.sender_user_id <> $1
	  AND m.read_at IS NULL
) AS unread ON TRUE
WHERE c.id = $2
  AND ($1 = c.participant_a_user_id OR $1 = c.participant_b_user_id)
`, userID, conversationID).Scan(&item.ID, &item.OpportunityID, &item.OpportunityTitle, &item.CompanyLegalName, &item.ParticipantUserID, &participantName, &participantAvatarURL, &item.ParticipantIsOnline, &participantLastSeenAt, &lastMessage, &lastMessageAt, &item.UnreadCount, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("chat conversation not found")
		}
		return nil, fmt.Errorf("get chat conversation: %w", err)
	}
	item.ParticipantName = participantName
	if participantAvatarURL.Valid {
		item.ParticipantAvatarURL = participantAvatarURL.String
	}
	if participantLastSeenAt.Valid {
		item.ParticipantLastSeenAt = &participantLastSeenAt.Time
	}
	if lastMessage.Valid {
		item.LastMessage = lastMessage.String
	}
	if lastMessageAt.Valid {
		item.LastMessageAt = lastMessageAt.Time
	}
	return &item, nil
}

func (r *Repository) ListChatConversations(userID string) ([]ChatConversation, error) {
	rows, err := r.db.QueryContext(context.Background(), `
SELECT
	c.id,
	COALESCE(c.opportunity_id::text, ''),
	COALESCE(o.title, ''),
	COALESCE(comp.legal_name, ''),
	CASE WHEN c.participant_a_user_id = $1 THEN c.participant_b_user_id ELSE c.participant_a_user_id END AS participant_user_id,
	u.display_name,
	COALESCE(u.avatar_url, ''),
	COALESCE(up.is_online, FALSE),
	up.last_seen_at,
	COALESCE(last_message.body, ''),
	last_message.created_at,
	COALESCE(unread.unread_count, 0),
	c.created_at,
	c.updated_at
FROM chat_conversations c
JOIN users u ON u.id = CASE WHEN c.participant_a_user_id = $1 THEN c.participant_b_user_id ELSE c.participant_a_user_id END
LEFT JOIN user_presence up ON up.user_id = u.id
LEFT JOIN opportunities o ON o.id = c.opportunity_id
LEFT JOIN companies comp ON comp.id = o.company_id
LEFT JOIN LATERAL (
	SELECT body, created_at
	FROM chat_messages
	WHERE conversation_id = c.id
	ORDER BY created_at DESC
	LIMIT 1
) AS last_message ON TRUE
LEFT JOIN LATERAL (
	SELECT COUNT(*) AS unread_count
	FROM chat_messages m
	WHERE m.conversation_id = c.id
	  AND m.sender_user_id <> $1
	  AND m.read_at IS NULL
) AS unread ON TRUE
WHERE c.participant_a_user_id = $1 OR c.participant_b_user_id = $1
ORDER BY COALESCE(last_message.created_at, c.updated_at) DESC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("list chat conversations: %w", err)
	}
	defer rows.Close()

	var result []ChatConversation
	for rows.Next() {
		var item ChatConversation
		var participantName string
		var participantAvatarURL sql.NullString
		var participantLastSeenAt sql.NullTime
		var lastMessage sql.NullString
		var lastMessageAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.OpportunityID, &item.OpportunityTitle, &item.CompanyLegalName, &item.ParticipantUserID, &participantName, &participantAvatarURL, &item.ParticipantIsOnline, &participantLastSeenAt, &lastMessage, &lastMessageAt, &item.UnreadCount, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan chat conversations: %w", err)
		}
		item.ParticipantName = participantName
		if participantAvatarURL.Valid {
			item.ParticipantAvatarURL = participantAvatarURL.String
		}
		if participantLastSeenAt.Valid {
			item.ParticipantLastSeenAt = &participantLastSeenAt.Time
		}
		if lastMessage.Valid {
			item.LastMessage = lastMessage.String
		}
		if lastMessageAt.Valid {
			item.LastMessageAt = lastMessageAt.Time
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) ListChatMessages(userID, conversationID string) ([]ChatMessage, error) {
	if _, err := r.GetChatConversation(userID, conversationID); err != nil {
		return nil, err
	}
	if _, err := r.MarkChatMessagesRead(userID, conversationID); err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(context.Background(), `
SELECT
	m.id,
	m.conversation_id,
	m.sender_user_id,
	u.display_name,
	COALESCE(u.avatar_url, ''),
	COALESCE(up.is_online, FALSE),
	m.body,
	(m.read_at IS NOT NULL),
	m.read_at,
	m.created_at
FROM chat_messages m
JOIN users u ON u.id = m.sender_user_id
LEFT JOIN user_presence up ON up.user_id = u.id
WHERE m.conversation_id = $1
ORDER BY m.created_at ASC
`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("list chat messages: %w", err)
	}
	defer rows.Close()

	var result []ChatMessage
	for rows.Next() {
		var item ChatMessage
		var readAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.ConversationID, &item.SenderUserID, &item.SenderName, &item.SenderAvatarURL, &item.SenderIsOnline, &item.Body, &item.IsRead, &readAt, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan chat messages: %w", err)
		}
		if readAt.Valid {
			item.ReadAt = &readAt.Time
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) CreateChatMessage(userID, conversationID, body string) (*ChatMessage, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, errors.New("message body is required")
	}
	if _, err := r.GetChatConversation(userID, conversationID); err != nil {
		return nil, err
	}
	item := &ChatMessage{
		ID:             uuid.NewString(),
		ConversationID: conversationID,
		SenderUserID:   userID,
		Body:           body,
		CreatedAt:      time.Now(),
	}
	if _, err := r.db.ExecContext(context.Background(), `
INSERT INTO chat_messages (id, conversation_id, sender_user_id, body, created_at)
VALUES ($1, $2, $3, $4, $5)
`, item.ID, item.ConversationID, item.SenderUserID, item.Body, item.CreatedAt); err != nil {
		return nil, fmt.Errorf("create chat message: %w", err)
	}
	if _, err := r.db.ExecContext(context.Background(), `
UPDATE chat_conversations
SET updated_at = $2
WHERE id = $1
`, conversationID, item.CreatedAt); err != nil {
		return nil, fmt.Errorf("update chat conversation timestamp: %w", err)
	}
	sender, err := r.getUserByID(context.Background(), userID)
	if err != nil {
		return nil, err
	}
	item.SenderName = sender.DisplayName
	item.SenderAvatarURL = sender.AvatarURL
	item.SenderIsOnline = sender.IsOnline
	return item, nil
}

func (r *Repository) MarkChatMessagesRead(userID, conversationID string) (int64, error) {
	if _, err := r.GetChatConversation(userID, conversationID); err != nil {
		return 0, err
	}
	result, err := r.db.ExecContext(context.Background(), `
UPDATE chat_messages
SET read_at = NOW()
WHERE conversation_id = $1
  AND sender_user_id <> $2
  AND read_at IS NULL
`, conversationID, userID)
	if err != nil {
		return 0, fmt.Errorf("mark chat messages read: %w", err)
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("count marked chat messages: %w", err)
	}
	return updated, nil
}
