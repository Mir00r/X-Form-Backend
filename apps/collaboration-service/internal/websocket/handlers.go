package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kamkaiz/x-form-backend/collaboration-service/internal/models"
	"go.uber.org/zap"
)

// Helper function to convert payload to specific type
func convertPayload(payload interface{}, target interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return nil
}

// registerEventHandlers registers all event handlers
func (h *Hub) registerEventHandlers() {
	h.eventHandlers[models.EventJoinForm] = &JoinFormHandler{hub: h}
	h.eventHandlers[models.EventLeaveForm] = &LeaveFormHandler{hub: h}
	h.eventHandlers[models.EventCursorUpdate] = &CursorUpdateHandler{hub: h}
	h.eventHandlers[models.EventQuestionUpdate] = &QuestionUpdateHandler{hub: h}
	h.eventHandlers[models.EventQuestionCreate] = &QuestionCreateHandler{hub: h}
	h.eventHandlers[models.EventQuestionDelete] = &QuestionDeleteHandler{hub: h}
	h.eventHandlers[models.EventPing] = &PingHandler{hub: h}
}

// JoinFormHandler handles join form events
type JoinFormHandler struct {
	hub *Hub
}

func (h *JoinFormHandler) Handle(ctx context.Context, client *Client, message *models.Message) error {
	var payload models.JoinFormPayload

	// Handle payload conversion
	payloadBytes, err := json.Marshal(message.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("invalid join form payload: %w", err)
	}

	// Validate form access
	if !h.hub.auth.CanAccessForm(client.User, payload.FormID) {
		return fmt.Errorf("access denied to form: %s", payload.FormID)
	}

	// Update client form ID
	client.FormID = payload.FormID

	// Join room
	if err := h.hub.joinRoom(payload.FormID, client.UserID, client.User); err != nil {
		return fmt.Errorf("failed to join room: %w", err)
	}

	// Get room info
	room, _ := h.hub.GetRoom(payload.FormID)

	// Create join response
	response := models.NewMessage(models.EventJoinFormResponse, &models.JoinFormResponsePayload{
		FormID:    payload.FormID,
		UserID:    client.UserID,
		Success:   true,
		RoomUsers: room.GetUserList(),
		Timestamp: time.Now(),
	})

	// Send response to client
	select {
	case client.send <- response:
	default:
		return fmt.Errorf("failed to send join response")
	}

	// Broadcast user joined to room
	joinedMessage := models.NewMessage(models.EventUserJoined, &models.UserJoinedPayload{
		FormID: payload.FormID,
		User:   client.User,
	})
	joinedMessage.FormID = payload.FormID

	h.hub.broadcast <- joinedMessage

	// Save user session
	session := &models.UserSession{
		UserID:    client.UserID,
		FormID:    payload.FormID,
		JoinedAt:  time.Now(),
		IsActive:  true,
		ClientID:  client.ID,
		UserAgent: client.UserAgent,
		IPAddress: client.IPAddress,
	}

	if err := h.hub.redis.SaveUserFormSession(ctx, client.UserID, payload.FormID, session); err != nil {
		h.hub.logger.Error("Failed to save user session", zap.Error(err))
	}

	h.hub.logger.Info("User joined form",
		zap.String("userID", client.UserID),
		zap.String("formID", payload.FormID),
		zap.String("userName", client.User.Name))

	return nil
}

// LeaveFormHandler handles leave form events
type LeaveFormHandler struct {
	hub *Hub
}

func (h *LeaveFormHandler) Handle(ctx context.Context, client *Client, message *models.Message) error {
	var payload models.LeaveFormPayload

	// Handle payload conversion
	payloadBytes, err := json.Marshal(message.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("invalid leave form payload: %w", err)
	}

	formID := payload.FormID
	if formID == "" {
		formID = client.FormID
	}

	if formID == "" {
		return fmt.Errorf("no form to leave")
	}

	// Remove user from room
	h.hub.removeUserFromRoom(formID, client.UserID)

	// Clear client form ID
	client.FormID = ""

	// Create leave response
	response := models.NewMessage(models.EventLeaveFormResponse, &models.LeaveFormResponsePayload{
		FormID:    formID,
		UserID:    client.UserID,
		Success:   true,
		Timestamp: time.Now(),
	})

	// Send response to client
	select {
	case client.send <- response:
	default:
		return fmt.Errorf("failed to send leave response")
	}

	// Broadcast user left to room
	leftMessage := models.NewMessage(models.EventUserLeft, &models.UserLeftPayload{
		FormID: formID,
		UserID: client.UserID,
	})
	leftMessage.FormID = formID

	h.hub.broadcast <- leftMessage

	// Remove user session
	if err := h.hub.redis.RemoveUserSession(ctx, client.UserID, formID); err != nil {
		h.hub.logger.Error("Failed to remove user session", zap.Error(err))
	}

	h.hub.logger.Info("User left form",
		zap.String("userID", client.UserID),
		zap.String("formID", formID),
		zap.String("userName", client.User.Name))

	return nil
}

// CursorUpdateHandler handles cursor update events
type CursorUpdateHandler struct {
	hub *Hub
}

func (h *CursorUpdateHandler) Handle(ctx context.Context, client *Client, message *models.Message) error {
	var payload models.CursorUpdatePayload
	if err := convertPayload(message.Payload, &payload); err != nil {
		return fmt.Errorf("invalid cursor update payload: %w", err)
	}

	// Validate form access
	if client.FormID == "" || client.FormID != payload.FormID {
		return fmt.Errorf("not joined to form or form mismatch")
	}

	// Update cursor position in Redis
	cursor := &models.CursorPosition{
		UserID:      client.UserID,
		FormID:      payload.FormID,
		QuestionID:  payload.Position.QuestionID,
		X:           payload.Position.X,
		Y:           payload.Position.Y,
		Section:     payload.Position.Section,
		LastUpdated: time.Now(),
	}

	if err := h.hub.redis.UpdateCursor(ctx, client.UserID, payload.FormID, cursor); err != nil {
		h.hub.logger.Error("Failed to update cursor in Redis", zap.Error(err))
	}

	// Broadcast cursor update to room (excluding sender)
	broadcastMessage := models.NewMessage(models.EventCursorUpdate, &models.CursorUpdatePayload{
		FormID:   payload.FormID,
		UserID:   client.UserID,
		Position: payload.Position,
		User:     client.User,
	})
	broadcastMessage.FormID = payload.FormID

	// Send to all users in room except sender
	h.hub.broadcastToRoomExceptUser(payload.FormID, client.UserID, broadcastMessage)

	return nil
}

// QuestionUpdateHandler handles question update events
type QuestionUpdateHandler struct {
	hub *Hub
}

func (h *QuestionUpdateHandler) Handle(ctx context.Context, client *Client, message *models.Message) error {
	var payload models.QuestionUpdatePayload
	if err := convertPayload(message.Payload, &payload); err != nil {
		return fmt.Errorf("invalid question update payload: %w", err)
	}

	// Validate form access and edit permissions
	if client.FormID == "" || client.FormID != payload.FormID {
		return fmt.Errorf("not joined to form or form mismatch")
	}

	if !h.hub.auth.CanEditForm(client.User, payload.FormID) {
		return fmt.Errorf("insufficient permissions to edit form")
	}

	// Save update to Redis for conflict resolution
	update := &models.QuestionUpdate{
		QuestionID: payload.QuestionID,
		FormID:     payload.FormID,
		UserID:     client.UserID,
		UpdateType: "update",
		Changes:    payload.Changes,
		Timestamp:  time.Now(),
		Version:    payload.Version,
	}

	if err := h.hub.redis.SaveQuestionUpdate(ctx, update); err != nil {
		h.hub.logger.Error("Failed to save question update", zap.Error(err))
	}

	// Broadcast update to room
	broadcastMessage := models.NewMessage(models.EventQuestionUpdate, &payload)
	broadcastMessage.FormID = payload.FormID

	h.hub.broadcast <- broadcastMessage

	// Send to Kafka for form service synchronization
	kafkaEvent := &models.KafkaEvent{
		Type:      "question.updated",
		FormID:    payload.FormID,
		UserID:    client.UserID,
		Data:      payload,
		Timestamp: time.Now(),
	}

	if err := h.hub.publishKafkaEvent(ctx, kafkaEvent); err != nil {
		h.hub.logger.Error("Failed to publish Kafka event", zap.Error(err))
	}

	h.hub.logger.Info("Question updated",
		zap.String("userID", client.UserID),
		zap.String("formID", payload.FormID),
		zap.String("questionID", payload.QuestionID))

	return nil
}

// QuestionCreateHandler handles question create events
type QuestionCreateHandler struct {
	hub *Hub
}

func (h *QuestionCreateHandler) Handle(ctx context.Context, client *Client, message *models.Message) error {
	var payload models.QuestionCreatePayload
	if err := convertPayload(message.Payload, &payload); err != nil {
		return fmt.Errorf("invalid question create payload: %w", err)
	}

	// Validate form access and edit permissions
	if client.FormID == "" || client.FormID != payload.FormID {
		return fmt.Errorf("not joined to form or form mismatch")
	}

	if !h.hub.auth.CanEditForm(client.User, payload.FormID) {
		return fmt.Errorf("insufficient permissions to edit form")
	}

	// Save creation to Redis
	update := &models.QuestionUpdate{
		QuestionID: payload.Question.ID,
		FormID:     payload.FormID,
		UserID:     client.UserID,
		UpdateType: "create",
		Changes:    payload.Question,
		Timestamp:  time.Now(),
	}

	if err := h.hub.redis.SaveQuestionUpdate(ctx, update); err != nil {
		h.hub.logger.Error("Failed to save question creation", zap.Error(err))
	}

	// Broadcast creation to room
	broadcastMessage := models.NewMessage(models.EventQuestionCreate, &payload)
	broadcastMessage.FormID = payload.FormID

	h.hub.broadcast <- broadcastMessage

	// Send to Kafka for form service synchronization
	kafkaEvent := &models.KafkaEvent{
		Type:      "question.created",
		FormID:    payload.FormID,
		UserID:    client.UserID,
		Data:      payload,
		Timestamp: time.Now(),
	}

	if err := h.hub.publishKafkaEvent(ctx, kafkaEvent); err != nil {
		h.hub.logger.Error("Failed to publish Kafka event", zap.Error(err))
	}

	h.hub.logger.Info("Question created",
		zap.String("userID", client.UserID),
		zap.String("formID", payload.FormID),
		zap.String("questionID", payload.Question.ID))

	return nil
}

// QuestionDeleteHandler handles question delete events
type QuestionDeleteHandler struct {
	hub *Hub
}

func (h *QuestionDeleteHandler) Handle(ctx context.Context, client *Client, message *models.Message) error {
	var payload models.QuestionDeletePayload
	if err := convertPayload(message.Payload, &payload); err != nil {
		return fmt.Errorf("invalid question delete payload: %w", err)
	}

	// Validate form access and edit permissions
	if client.FormID == "" || client.FormID != payload.FormID {
		return fmt.Errorf("not joined to form or form mismatch")
	}

	if !h.hub.auth.CanEditForm(client.User, payload.FormID) {
		return fmt.Errorf("insufficient permissions to edit form")
	}

	// Save deletion to Redis
	update := &models.QuestionUpdate{
		QuestionID: payload.QuestionID,
		FormID:     payload.FormID,
		UserID:     client.UserID,
		UpdateType: "delete",
		Timestamp:  time.Now(),
	}

	if err := h.hub.redis.SaveQuestionUpdate(ctx, update); err != nil {
		h.hub.logger.Error("Failed to save question deletion", zap.Error(err))
	}

	// Broadcast deletion to room
	broadcastMessage := models.NewMessage(models.EventQuestionDelete, &payload)
	broadcastMessage.FormID = payload.FormID

	h.hub.broadcast <- broadcastMessage

	// Send to Kafka for form service synchronization
	kafkaEvent := &models.KafkaEvent{
		Type:      "question.deleted",
		FormID:    payload.FormID,
		UserID:    client.UserID,
		Data:      payload,
		Timestamp: time.Now(),
	}

	if err := h.hub.publishKafkaEvent(ctx, kafkaEvent); err != nil {
		h.hub.logger.Error("Failed to publish Kafka event", zap.Error(err))
	}

	h.hub.logger.Info("Question deleted",
		zap.String("userID", client.UserID),
		zap.String("formID", payload.FormID),
		zap.String("questionID", payload.QuestionID))

	return nil
}

// PingHandler handles ping events
type PingHandler struct {
	hub *Hub
}

func (h *PingHandler) Handle(ctx context.Context, client *Client, message *models.Message) error {
	// Update last ping time
	client.LastPing = time.Now()

	// Send pong response
	pongMessage := models.NewMessage(models.EventPong, &models.PongPayload{
		Timestamp: time.Now(),
	})

	select {
	case client.send <- pongMessage:
	default:
		return fmt.Errorf("failed to send pong response")
	}

	return nil
}

// Helper methods

// broadcastToRoomExceptUser broadcasts a message to all users in a room except the specified user
func (h *Hub) broadcastToRoomExceptUser(formID, excludeUserID string, message *models.Message) {
	h.mu.RLock()
	room, exists := h.rooms[formID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	for userID := range room.Users {
		if userID != excludeUserID {
			h.sendToUser(userID, message)
		}
	}
}

// publishKafkaEvent publishes an event to Kafka (placeholder - implement with actual Kafka client)
func (h *Hub) publishKafkaEvent(ctx context.Context, event *models.KafkaEvent) error {
	// TODO: Implement Kafka producer
	// This is a placeholder for Kafka integration
	h.logger.Info("Publishing Kafka event",
		zap.String("type", event.Type),
		zap.String("formID", event.FormID),
		zap.String("userID", event.UserID))

	return nil
}
