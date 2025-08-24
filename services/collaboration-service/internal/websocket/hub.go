package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/kamkaiz/x-form-backend/collaboration-service/internal/auth"
	"github.com/kamkaiz/x-form-backend/collaboration-service/internal/config"
	"github.com/kamkaiz/x-form-backend/collaboration-service/internal/models"
	redisService "github.com/kamkaiz/x-form-backend/collaboration-service/internal/redis"
	"go.uber.org/zap"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Inbound messages from the clients
	broadcast chan *models.Message

	// Room management
	rooms map[string]*models.Room

	// User connections mapping
	userConnections map[string][]*Client

	// Redis service for persistence
	redis *redisService.Service

	// Auth service
	auth *auth.Service

	// Configuration
	config *config.WebSocketConfig

	// Logger
	logger *zap.Logger

	// Metrics
	metrics *Metrics

	// Mutex for thread safety
	mu sync.RWMutex

	// Rate limiting
	rateLimiter *RateLimiter

	// Event handlers
	eventHandlers map[models.EventType]EventHandler
}

// Client represents a WebSocket client
type Client struct {
	// WebSocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan *models.Message

	// Hub reference
	hub *Hub

	// Client metadata
	ID     string
	UserID string
	User   *models.User
	FormID string

	// Connection info
	ConnectedAt time.Time
	LastPing    time.Time
	IsActive    bool
	UserAgent   string
	IPAddress   string

	// Rate limiting
	rateLimitInfo *models.RateLimitInfo

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// EventHandler defines the interface for handling WebSocket events
type EventHandler interface {
	Handle(ctx context.Context, client *Client, message *models.Message) error
}

// Metrics tracks WebSocket metrics
type Metrics struct {
	TotalConnections  int64
	ActiveConnections int64
	TotalRooms        int64
	ActiveRooms       int64
	MessagesPerSecond int64
	ErrorsPerSecond   int64
	mu                sync.RWMutex
}

// RateLimiter handles rate limiting for WebSocket connections
type RateLimiter struct {
	redis  *redisService.Service
	config *config.WebSocketConfig
}

// NewHub creates a new WebSocket hub
func NewHub(redis *redisService.Service, authService *auth.Service, cfg *config.WebSocketConfig, logger *zap.Logger) *Hub {
	hub := &Hub{
		clients:         make(map[*Client]bool),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		broadcast:       make(chan *models.Message),
		rooms:           make(map[string]*models.Room),
		userConnections: make(map[string][]*Client),
		redis:           redis,
		auth:            authService,
		config:          cfg,
		logger:          logger,
		metrics:         &Metrics{},
		rateLimiter:     NewRateLimiter(redis, cfg),
		eventHandlers:   make(map[models.EventType]EventHandler),
	}

	// Register event handlers
	hub.registerEventHandlers()

	return hub
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redis *redisService.Service, config *config.WebSocketConfig) *RateLimiter {
	return &RateLimiter{
		redis:  redis,
		config: config,
	}
}

// Run starts the hub and handles client connections
func (h *Hub) Run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)

		case <-ticker.C:
			h.cleanup()

		case <-ctx.Done():
			h.logger.Info("WebSocket hub shutting down")
			return
		}
	}
}

// ServeWS handles WebSocket requests from clients
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Configure WebSocket upgrader
	upgrader := websocket.Upgrader{
		ReadBufferSize:    h.config.ReadBufferSize,
		WriteBufferSize:   h.config.WriteBufferSize,
		CheckOrigin:       h.checkOrigin,
		EnableCompression: h.config.EnableCompression,
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	// Authenticate the connection
	user, err := h.authenticateConnection(r)
	if err != nil {
		h.logger.Error("Authentication failed", zap.Error(err))
		conn.Close()
		return
	}

	// Create client
	client := h.createClient(conn, user, r)

	// Register client with hub
	h.register <- client

	// Start client goroutines
	go client.writePump()
	go client.readPump()
}

// authenticateConnection authenticates a WebSocket connection
func (h *Hub) authenticateConnection(r *http.Request) (*models.User, error) {
	// Extract token from query parameters or headers
	token := auth.ExtractTokenFromQuery(r.URL.Query())
	if token == "" {
		token = auth.ExtractTokenFromHeader(r.Header.Get("Authorization"))
	}

	if token == "" {
		return nil, fmt.Errorf("no authentication token provided")
	}

	// Validate token
	claims, err := h.auth.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Create user from claims
	user := h.auth.CreateUser(claims)
	return user, nil
}

// createClient creates a new WebSocket client
func (h *Hub) createClient(conn *websocket.Conn, user *models.User, r *http.Request) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		conn:        conn,
		send:        make(chan *models.Message, 256),
		hub:         h,
		ID:          uuid.New().String(),
		UserID:      user.ID,
		User:        user,
		ConnectedAt: time.Now(),
		LastPing:    time.Now(),
		IsActive:    true,
		UserAgent:   r.Header.Get("User-Agent"),
		IPAddress:   r.RemoteAddr,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Set connection timeouts
	conn.SetReadLimit(h.config.MaxMessageSize)
	conn.SetReadDeadline(time.Now().Add(h.config.PongWait))
	conn.SetPongHandler(func(string) error {
		client.LastPing = time.Now()
		conn.SetReadDeadline(time.Now().Add(h.config.PongWait))
		return nil
	})

	return client
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Add to clients map
	h.clients[client] = true

	// Add to user connections
	if h.userConnections[client.UserID] == nil {
		h.userConnections[client.UserID] = make([]*Client, 0)
	}
	h.userConnections[client.UserID] = append(h.userConnections[client.UserID], client)

	// Update metrics
	h.metrics.mu.Lock()
	h.metrics.TotalConnections++
	h.metrics.ActiveConnections++
	h.metrics.mu.Unlock()

	// Save connection to Redis
	connection := &models.Connection{
		ID:        client.ID,
		UserID:    client.UserID,
		Connected: client.ConnectedAt,
		LastPing:  client.LastPing,
		IsActive:  client.IsActive,
		UserAgent: client.UserAgent,
		IPAddress: client.IPAddress,
	}

	if err := h.redis.SaveConnection(context.Background(), client.ID, connection); err != nil {
		h.logger.Error("Failed to save connection to Redis", zap.Error(err))
	}

	h.logger.Info("Client connected",
		zap.String("clientID", client.ID),
		zap.String("userID", client.UserID),
		zap.String("userName", client.User.Name))
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		// Remove from clients map
		delete(h.clients, client)

		// Remove from user connections
		userConns := h.userConnections[client.UserID]
		for i, conn := range userConns {
			if conn.ID == client.ID {
				h.userConnections[client.UserID] = append(userConns[:i], userConns[i+1:]...)
				break
			}
		}

		// If no more connections for user, remove from map
		if len(h.userConnections[client.UserID]) == 0 {
			delete(h.userConnections, client.UserID)
		}

		// Close send channel
		close(client.send)

		// Cancel client context
		client.cancel()

		// Remove from current room
		if client.FormID != "" {
			h.removeUserFromRoom(client.FormID, client.UserID)
		}

		// Update metrics
		h.metrics.mu.Lock()
		h.metrics.ActiveConnections--
		h.metrics.mu.Unlock()

		// Remove connection from Redis
		if err := h.redis.DeleteConnection(context.Background(), client.ID); err != nil {
			h.logger.Error("Failed to delete connection from Redis", zap.Error(err))
		}

		h.logger.Info("Client disconnected",
			zap.String("clientID", client.ID),
			zap.String("userID", client.UserID))
	}
}

// broadcastMessage broadcasts a message to relevant clients
func (h *Hub) broadcastMessage(message *models.Message) {
	if message.FormID != "" {
		// Broadcast to room
		h.broadcastToRoom(message.FormID, message)
	} else if message.UserID != "" {
		// Send to specific user
		h.sendToUser(message.UserID, message)
	} else {
		// Broadcast to all clients
		h.broadcastToAll(message)
	}
}

// broadcastToRoom broadcasts a message to all clients in a room
func (h *Hub) broadcastToRoom(formID string, message *models.Message) {
	h.mu.RLock()
	room, exists := h.rooms[formID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	for userID := range room.Users {
		h.sendToUser(userID, message)
	}
}

// sendToUser sends a message to all connections of a specific user
func (h *Hub) sendToUser(userID string, message *models.Message) {
	h.mu.RLock()
	connections := h.userConnections[userID]
	h.mu.RUnlock()

	for _, client := range connections {
		select {
		case client.send <- message:
		default:
			// Client's send channel is full, close it
			h.unregisterClient(client)
		}
	}
}

// broadcastToAll broadcasts a message to all connected clients
func (h *Hub) broadcastToAll(message *models.Message) {
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.send <- message:
		default:
			h.unregisterClient(client)
		}
	}
}

// readPump handles reading messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			// Read message from WebSocket
			_, messageData, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.hub.logger.Error("WebSocket error", zap.Error(err))
				}
				return
			}

			// Parse message
			var message models.Message
			if err := json.Unmarshal(messageData, &message); err != nil {
				c.sendError("INVALID_MESSAGE", "Failed to parse message")
				continue
			}

			// Check rate limit
			if err := c.checkRateLimit(); err != nil {
				c.sendError("RATE_LIMIT", "Rate limit exceeded")
				continue
			}

			// Set message metadata
			message.UserID = c.UserID
			message.Timestamp = time.Now()
			message.MessageID = uuid.New().String()

			// Handle message
			if err := c.handleMessage(&message); err != nil {
				c.hub.logger.Error("Failed to handle message", zap.Error(err))
				c.sendError("HANDLER_ERROR", "Failed to process message")
			}
		}
	}
}

// writePump handles writing messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(c.hub.config.PingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(c.hub.config.WriteWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send message
			if err := c.sendMessage(message); err != nil {
				c.hub.logger.Error("Failed to send message", zap.Error(err))
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(c.hub.config.WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.ctx.Done():
			return
		}
	}
}

// sendMessage sends a message through the WebSocket connection
func (c *Client) sendMessage(message *models.Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// sendError sends an error message to the client
func (c *Client) sendError(code, message string) {
	errorMsg := models.NewMessage(models.EventError, &models.ErrorPayload{
		Code:    code,
		Message: message,
	})
	errorMsg.UserID = c.UserID

	select {
	case c.send <- errorMsg:
	default:
		// Channel is full, drop the error message
	}
}

// handleMessage handles incoming messages from clients
func (c *Client) handleMessage(message *models.Message) error {
	// Get event handler
	handler, exists := c.hub.eventHandlers[message.Type]
	if !exists {
		return fmt.Errorf("no handler for event type: %s", message.Type)
	}

	// Handle the message
	return handler.Handle(c.ctx, c, message)
}

// checkRateLimit checks if the client is rate limited
func (c *Client) checkRateLimit() error {
	rateLimitInfo, err := c.hub.rateLimiter.CheckRateLimit(
		context.Background(),
		c.UserID,
		c.hub.config.MessageRateLimit,
		c.hub.config.RateLimitWindow,
	)
	if err != nil {
		return err
	}

	c.rateLimitInfo = rateLimitInfo

	if !rateLimitInfo.IsAllowed() {
		return fmt.Errorf("rate limit exceeded")
	}

	return nil
}

// checkOrigin checks the origin of WebSocket connections
func (h *Hub) checkOrigin(r *http.Request) bool {
	if !h.config.CheckOrigin {
		return true
	}

	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}

	// Check against allowed origins
	for _, allowedOrigin := range h.config.AllowedOrigins {
		if origin == allowedOrigin {
			return true
		}
	}

	return false
}

// cleanup performs periodic cleanup tasks
func (h *Hub) cleanup() {
	// Clean up inactive rooms
	h.cleanupInactiveRooms()

	// Update metrics
	h.updateMetrics()

	// Clean up Redis data
	if err := h.redis.CleanupExpiredData(context.Background()); err != nil {
		h.logger.Error("Failed to cleanup Redis data", zap.Error(err))
	}
}

// cleanupInactiveRooms removes rooms with no active users
func (h *Hub) cleanupInactiveRooms() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for formID, room := range h.rooms {
		if len(room.Users) == 0 && time.Since(room.UpdatedAt) > time.Hour {
			delete(h.rooms, formID)
			h.redis.DeleteRoom(context.Background(), formID)
		}
	}
}

// updateMetrics updates WebSocket metrics
func (h *Hub) updateMetrics() {
	h.metrics.mu.Lock()
	defer h.metrics.mu.Unlock()

	h.metrics.TotalRooms = int64(len(h.rooms))

	activeRooms := int64(0)
	for _, room := range h.rooms {
		if len(room.Users) > 0 {
			activeRooms++
		}
	}
	h.metrics.ActiveRooms = activeRooms
}

// GetMetrics returns current WebSocket metrics
func (h *Hub) GetMetrics() *Metrics {
	h.metrics.mu.RLock()
	defer h.metrics.mu.RUnlock()

	return &Metrics{
		TotalConnections:  h.metrics.TotalConnections,
		ActiveConnections: h.metrics.ActiveConnections,
		TotalRooms:        h.metrics.TotalRooms,
		ActiveRooms:       h.metrics.ActiveRooms,
		MessagesPerSecond: h.metrics.MessagesPerSecond,
		ErrorsPerSecond:   h.metrics.ErrorsPerSecond,
	}
}

// CheckRateLimit checks rate limit for a user
func (rl *RateLimiter) CheckRateLimit(ctx context.Context, userID string, limit int, window time.Duration) (*models.RateLimitInfo, error) {
	return rl.redis.CheckRateLimit(ctx, userID, limit, window)
}

// Room management methods

// joinRoom adds a user to a room
func (h *Hub) joinRoom(formID, userID string, user *models.User) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Get or create room
	room, exists := h.rooms[formID]
	if !exists {
		room = models.NewRoom(formID, h.config.MaxUsersPerRoom)
		h.rooms[formID] = room
	}

	// Add user to room
	if !room.AddUser(user) {
		return fmt.Errorf("room is full")
	}

	// Save room to Redis
	if err := h.redis.SaveRoom(context.Background(), room); err != nil {
		h.logger.Error("Failed to save room to Redis", zap.Error(err))
	}

	// Add user to Redis room users set
	if err := h.redis.AddUserToRoom(context.Background(), formID, userID); err != nil {
		h.logger.Error("Failed to add user to room in Redis", zap.Error(err))
	}

	return nil
}

// removeUserFromRoom removes a user from a room
func (h *Hub) removeUserFromRoom(formID, userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.rooms[formID]
	if !exists {
		return
	}

	// Remove user from room
	room.RemoveUser(userID)

	// Save room to Redis
	if err := h.redis.SaveRoom(context.Background(), room); err != nil {
		h.logger.Error("Failed to save room to Redis", zap.Error(err))
	}

	// Remove user from Redis room users set
	if err := h.redis.RemoveUserFromRoom(context.Background(), formID, userID); err != nil {
		h.logger.Error("Failed to remove user from room in Redis", zap.Error(err))
	}

	// Delete room if empty
	if len(room.Users) == 0 {
		delete(h.rooms, formID)
		if err := h.redis.DeleteRoom(context.Background(), formID); err != nil {
			h.logger.Error("Failed to delete room from Redis", zap.Error(err))
		}
	}
}

// GetRoom returns a room by form ID
func (h *Hub) GetRoom(formID string) (*models.Room, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, exists := h.rooms[formID]
	return room, exists
}

// GetActiveRooms returns all active rooms
func (h *Hub) GetActiveRooms() []*models.Room {
	h.mu.RLock()
	defer h.mu.RUnlock()

	rooms := make([]*models.Room, 0, len(h.rooms))
	for _, room := range h.rooms {
		if len(room.Users) > 0 {
			rooms = append(rooms, room)
		}
	}

	return rooms
}
