package handlers

import (
	"strings"
	"sync"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"

	"tramplin/internal/authjwt"
	"tramplin/internal/dto"
	chatservice "tramplin/internal/service/chat"
)

type ChatHandler struct {
	service *chatservice.Service
	jwt     *authjwt.Manager
	hub     *chatHub
}

type chatHub struct {
	mu    sync.Mutex
	rooms map[string]map[*websocket.Conn]struct{}
}

func newChatHub() *chatHub {
	return &chatHub{rooms: make(map[string]map[*websocket.Conn]struct{})}
}

func NewChatHandler(service *chatservice.Service, jwtManager *authjwt.Manager) *ChatHandler {
	return &ChatHandler{service: service, jwt: jwtManager, hub: newChatHub()}
}

func (h *chatHub) join(conversationID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[conversationID] == nil {
		h.rooms[conversationID] = make(map[*websocket.Conn]struct{})
	}
	h.rooms[conversationID][conn] = struct{}{}
}

func (h *chatHub) leave(conversationID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[conversationID] == nil {
		return
	}
	delete(h.rooms[conversationID], conn)
	if len(h.rooms[conversationID]) == 0 {
		delete(h.rooms, conversationID)
	}
}

func (h *chatHub) broadcast(conversationID string, payload any) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.rooms[conversationID] {
		if err := conn.WriteJSON(payload); err != nil {
			_ = conn.Close()
			delete(h.rooms[conversationID], conn)
		}
	}
	if len(h.rooms[conversationID]) == 0 {
		delete(h.rooms, conversationID)
	}
}

func (h *ChatHandler) CreateConversation(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input dto.ChatConversationInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.CreateConversation(userID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}

func (h *ChatHandler) ListConversations(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	data, err := h.service.ListConversations(userID)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

func (h *ChatHandler) ListMessages(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	data, err := h.service.ListMessages(userID, c.Params("id"))
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

func (h *ChatHandler) CreateMessage(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input dto.ChatMessageInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.CreateMessage(userID, c.Params("id"), input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	h.hub.broadcast(c.Params("id"), data)
	return respond(c, fiber.StatusCreated, data)
}

func (h *ChatHandler) WebSocket(c *fiber.Ctx) error {
	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		return fail(c, fiber.StatusUnauthorized, fiber.NewError(fiber.StatusUnauthorized, "token is required"))
	}
	if _, err := h.jwt.Parse(token); err != nil {
		return fail(c, fiber.StatusUnauthorized, fiber.NewError(fiber.StatusUnauthorized, "invalid token"))
	}
	conversationID := strings.TrimSpace(c.Query("conversation_id"))
	if conversationID == "" {
		return fail(c, fiber.StatusBadRequest, fiber.NewError(fiber.StatusBadRequest, "conversation_id is required"))
	}
	if !websocket.FastHTTPIsWebSocketUpgrade(c.Context()) {
		return fail(c, fiber.StatusUpgradeRequired, fiber.NewError(fiber.StatusUpgradeRequired, "websocket upgrade required"))
	}

	claims, err := h.jwt.Parse(token)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, fiber.NewError(fiber.StatusUnauthorized, "invalid token"))
	}
	if _, err := h.service.GetConversation(claims.UserID, conversationID); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}

	upgrader := websocket.FastHTTPUpgrader{
		CheckOrigin: func(_ *fasthttp.RequestCtx) bool { return true },
	}
	return upgrader.Upgrade(c.Context(), func(conn *websocket.Conn) {
		h.hub.join(conversationID, conn)
		defer h.hub.leave(conversationID, conn)
		defer conn.Close()

		for {
			var input dto.ChatMessageInput
			if err := conn.ReadJSON(&input); err != nil {
				return
			}
			message, err := h.service.CreateMessage(claims.UserID, conversationID, input)
			if err != nil {
				_ = conn.WriteJSON(map[string]any{"type": "error", "error": err.Error()})
				continue
			}
			h.hub.broadcast(conversationID, message)
		}
	})
}
