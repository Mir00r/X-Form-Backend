package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kamkaiz/x-form-backend/collaboration-service/internal/config"
	"github.com/kamkaiz/x-form-backend/collaboration-service/internal/models"
)

// Service wraps Redis client with application-specific methods
type Service struct {
	client    *redis.Client
	config    *config.RedisConfig
	keyPrefix string
}

// NewService creates a new Redis service
func NewService(cfg *config.RedisConfig) (*Service, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Service{
		client:    rdb,
		config:    cfg,
		keyPrefix: "collaboration-service",
	}, nil
}

// Close closes the Redis connection
func (s *Service) Close() error {
	return s.client.Close()
}

// Room management methods

// SaveRoom saves a room to Redis
func (s *Service) SaveRoom(ctx context.Context, room *models.Room) error {
	key := s.getRoomKey(room.FormID)
	data, err := json.Marshal(room)
	if err != nil {
		return fmt.Errorf("failed to marshal room: %w", err)
	}

	return s.client.Set(ctx, key, data, time.Hour).Err()
}

// GetRoom retrieves a room from Redis
func (s *Service) GetRoom(ctx context.Context, formID string) (*models.Room, error) {
	key := s.getRoomKey(formID)
	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Room doesn't exist
		}
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	var room models.Room
	if err := json.Unmarshal([]byte(data), &room); err != nil {
		return nil, fmt.Errorf("failed to unmarshal room: %w", err)
	}

	return &room, nil
}

// DeleteRoom deletes a room from Redis
func (s *Service) DeleteRoom(ctx context.Context, formID string) error {
	key := s.getRoomKey(formID)
	return s.client.Del(ctx, key).Err()
}

// GetActiveRooms returns all active rooms
func (s *Service) GetActiveRooms(ctx context.Context) ([]*models.Room, error) {
	pattern := s.getRoomKey("*")
	keys, err := s.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get room keys: %w", err)
	}

	var rooms []*models.Room
	for _, key := range keys {
		data, err := s.client.Get(ctx, key).Result()
		if err != nil {
			continue // Skip failed reads
		}

		var room models.Room
		if err := json.Unmarshal([]byte(data), &room); err != nil {
			continue // Skip invalid data
		}

		if room.IsActive {
			rooms = append(rooms, &room)
		}
	}

	return rooms, nil
}

// User session management

// SaveUserSession saves user session data
func (s *Service) SaveUserSession(ctx context.Context, userID string, session *models.SessionData) error {
	key := s.getUserSessionKey(userID)
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	expiration := time.Until(session.ExpiresAt)
	if expiration <= 0 {
		expiration = time.Hour // Default expiration
	}

	return s.client.Set(ctx, key, data, expiration).Err()
}

// GetUserSession retrieves user session data
func (s *Service) GetUserSession(ctx context.Context, userID string) (*models.SessionData, error) {
	key := s.getUserSessionKey(userID)
	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Session doesn't exist
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session models.SessionData
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// DeleteUserSession deletes user session
func (s *Service) DeleteUserSession(ctx context.Context, userID string) error {
	key := s.getUserSessionKey(userID)
	return s.client.Del(ctx, key).Err()
}

// AddUserToRoom adds a user to a room's user list
func (s *Service) AddUserToRoom(ctx context.Context, formID, userID string) error {
	key := s.getRoomUsersKey(formID)
	return s.client.SAdd(ctx, key, userID).Err()
}

// RemoveUserFromRoom removes a user from a room's user list
func (s *Service) RemoveUserFromRoom(ctx context.Context, formID, userID string) error {
	key := s.getRoomUsersKey(formID)
	return s.client.SRem(ctx, key, userID).Err()
}

// GetRoomUsers returns all users in a room
func (s *Service) GetRoomUsers(ctx context.Context, formID string) ([]string, error) {
	key := s.getRoomUsersKey(formID)
	return s.client.SMembers(ctx, key).Result()
}

// Connection management

// SaveConnection saves connection data
func (s *Service) SaveConnection(ctx context.Context, connID string, conn *models.Connection) error {
	key := s.getConnectionKey(connID)
	data, err := json.Marshal(conn)
	if err != nil {
		return fmt.Errorf("failed to marshal connection: %w", err)
	}

	return s.client.Set(ctx, key, data, time.Hour).Err()
}

// GetConnection retrieves connection data
func (s *Service) GetConnection(ctx context.Context, connID string) (*models.Connection, error) {
	key := s.getConnectionKey(connID)
	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	var conn models.Connection
	if err := json.Unmarshal([]byte(data), &conn); err != nil {
		return nil, fmt.Errorf("failed to unmarshal connection: %w", err)
	}

	return &conn, nil
}

// DeleteConnection deletes connection data
func (s *Service) DeleteConnection(ctx context.Context, connID string) error {
	key := s.getConnectionKey(connID)
	return s.client.Del(ctx, key).Err()
}

// Rate limiting

// CheckRateLimit checks and updates rate limit for a user
func (s *Service) CheckRateLimit(ctx context.Context, userID string, limit int, window time.Duration) (*models.RateLimitInfo, error) {
	key := s.getRateLimitKey(userID)

	// Use Redis pipeline for atomic operations
	pipe := s.client.Pipeline()

	// Increment counter
	incrCmd := pipe.Incr(ctx, key)

	// Set expiration if key is new
	pipe.Expire(ctx, key, window)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to execute rate limit pipeline: %w", err)
	}

	// Get current count
	currentCount := int(incrCmd.Val())

	// Calculate reset time
	ttl, _ := s.client.TTL(ctx, key).Result()
	resetTime := time.Now().Add(ttl)

	rateLimitInfo := &models.RateLimitInfo{
		UserID:    userID,
		Key:       key,
		Limit:     limit,
		Count:     currentCount,
		Window:    window,
		ResetTime: resetTime,
		Blocked:   currentCount > limit,
	}

	return rateLimitInfo, nil
}

// ResetRateLimit resets rate limit for a user
func (s *Service) ResetRateLimit(ctx context.Context, userID string) error {
	key := s.getRateLimitKey(userID)
	return s.client.Del(ctx, key).Err()
}

// Pub/Sub methods

// PublishMessage publishes a message to a channel
func (s *Service) PublishMessage(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return s.client.Publish(ctx, channel, data).Err()
}

// SubscribeToChannel subscribes to a Redis channel
func (s *Service) SubscribeToChannel(ctx context.Context, channels ...string) *redis.PubSub {
	return s.client.Subscribe(ctx, channels...)
}

// PublishToRoom publishes a message to all users in a room
func (s *Service) PublishToRoom(ctx context.Context, formID string, message *models.Message) error {
	channel := s.getRoomChannelKey(formID)
	return s.PublishMessage(ctx, channel, message)
}

// PublishToUser publishes a message to a specific user
func (s *Service) PublishToUser(ctx context.Context, userID string, message *models.Message) error {
	channel := s.getUserChannelKey(userID)
	return s.PublishMessage(ctx, channel, message)
}

// Cursor management

// SaveCursor saves cursor position for a user in a room
func (s *Service) SaveCursor(ctx context.Context, formID, userID string, cursor *models.Cursor) error {
	key := s.getCursorKey(formID, userID)
	data, err := json.Marshal(cursor)
	if err != nil {
		return fmt.Errorf("failed to marshal cursor: %w", err)
	}

	return s.client.Set(ctx, key, data, time.Minute*5).Err()
}

// GetCursor retrieves cursor position for a user in a room
func (s *Service) GetCursor(ctx context.Context, formID, userID string) (*models.Cursor, error) {
	key := s.getCursorKey(formID, userID)
	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get cursor: %w", err)
	}

	var cursor models.Cursor
	if err := json.Unmarshal([]byte(data), &cursor); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cursor: %w", err)
	}

	return &cursor, nil
}

// GetRoomCursors retrieves all cursors in a room
func (s *Service) GetRoomCursors(ctx context.Context, formID string) (map[string]*models.Cursor, error) {
	pattern := s.getCursorKey(formID, "*")
	keys, err := s.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get cursor keys: %w", err)
	}

	cursors := make(map[string]*models.Cursor)
	for _, key := range keys {
		data, err := s.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var cursor models.Cursor
		if err := json.Unmarshal([]byte(data), &cursor); err != nil {
			continue
		}

		cursors[cursor.UserID] = &cursor
	}

	return cursors, nil
}

// DeleteCursor deletes cursor for a user in a room
func (s *Service) DeleteCursor(ctx context.Context, formID, userID string) error {
	key := s.getCursorKey(formID, userID)
	return s.client.Del(ctx, key).Err()
}

// Metrics and monitoring

// IncrementMetric increments a metric counter
func (s *Service) IncrementMetric(ctx context.Context, metric string) error {
	key := s.getMetricKey(metric)
	return s.client.Incr(ctx, key).Err()
}

// SetMetric sets a metric value
func (s *Service) SetMetric(ctx context.Context, metric string, value int64) error {
	key := s.getMetricKey(metric)
	return s.client.Set(ctx, key, value, time.Hour).Err()
}

// GetMetric gets a metric value
func (s *Service) GetMetric(ctx context.Context, metric string) (int64, error) {
	key := s.getMetricKey(metric)
	return s.client.Get(ctx, key).Int64()
}

// Health check

// Ping checks Redis connectivity
func (s *Service) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

// GetStats returns Redis stats
func (s *Service) GetStats(ctx context.Context) (map[string]string, error) {
	info, err := s.client.Info(ctx).Result()
	if err != nil {
		return nil, err
	}

	// Parse info string into map
	stats := make(map[string]string)
	// Basic parsing - in production you might want more sophisticated parsing
	stats["info"] = info

	return stats, nil
}

// Key generation methods

func (s *Service) getRoomKey(formID string) string {
	return fmt.Sprintf("collaboration:room:%s", formID)
}

func (s *Service) getUserSessionKey(userID string) string {
	return fmt.Sprintf("collaboration:session:%s", userID)
}

func (s *Service) getRoomUsersKey(formID string) string {
	return fmt.Sprintf("collaboration:room:%s:users", formID)
}

func (s *Service) getConnectionKey(connID string) string {
	return fmt.Sprintf("collaboration:connection:%s", connID)
}

func (s *Service) getRateLimitKey(userID string) string {
	return fmt.Sprintf("collaboration:ratelimit:%s", userID)
}

func (s *Service) getRoomChannelKey(formID string) string {
	return fmt.Sprintf("collaboration:channel:room:%s", formID)
}

func (s *Service) getUserChannelKey(userID string) string {
	return fmt.Sprintf("collaboration:channel:user:%s", userID)
}

func (s *Service) getCursorKey(formID, userID string) string {
	return fmt.Sprintf("collaboration:cursor:%s:%s", formID, userID)
}

func (s *Service) getMetricKey(metric string) string {
	return fmt.Sprintf("collaboration:metrics:%s", metric)
}

// Batch operations

// BatchSaveRooms saves multiple rooms
func (s *Service) BatchSaveRooms(ctx context.Context, rooms []*models.Room) error {
	pipe := s.client.Pipeline()

	for _, room := range rooms {
		key := s.getRoomKey(room.FormID)
		data, err := json.Marshal(room)
		if err != nil {
			continue
		}
		pipe.Set(ctx, key, data, time.Hour)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// BatchDeleteKeys deletes multiple keys
func (s *Service) BatchDeleteKeys(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	return s.client.Del(ctx, keys...).Err()
}

// CleanupExpiredData removes expired data (called periodically)
func (s *Service) CleanupExpiredData(ctx context.Context) error {
	// This would implement cleanup logic for expired sessions, connections, etc.
	// For now, Redis handles TTL automatically, but we might want additional cleanup

	// Example: Clean up inactive rooms
	rooms, err := s.GetActiveRooms(ctx)
	if err != nil {
		return err
	}

	var keysToDelete []string
	for _, room := range rooms {
		// If room has no users and is older than threshold, mark for deletion
		if len(room.Users) == 0 && time.Since(room.UpdatedAt) > time.Hour {
			keysToDelete = append(keysToDelete, s.getRoomKey(room.FormID))
		}
	}

	return s.BatchDeleteKeys(ctx, keysToDelete)
}

// Additional missing methods needed by handlers

// SaveUserFormSession saves user session for a specific form
func (s *Service) SaveUserFormSession(ctx context.Context, userID, formID string, session *models.UserSession) error {
	key := s.getUserFormSessionKey(userID, formID)
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal user session: %w", err)
	}

	return s.client.Set(ctx, key, data, time.Hour*2).Err()
}

// RemoveUserSession removes user session for a specific form
func (s *Service) RemoveUserSession(ctx context.Context, userID, formID string) error {
	key := s.getUserFormSessionKey(userID, formID)
	return s.client.Del(ctx, key).Err()
}

// UpdateCursor updates cursor position for a user in a form
func (s *Service) UpdateCursor(ctx context.Context, userID, formID string, cursor *models.CursorPosition) error {
	key := s.getCursorPositionKey(formID, userID)
	data, err := json.Marshal(cursor)
	if err != nil {
		return fmt.Errorf("failed to marshal cursor position: %w", err)
	}

	return s.client.Set(ctx, key, data, time.Minute*5).Err()
}

// SaveQuestionUpdate saves question update for conflict resolution
func (s *Service) SaveQuestionUpdate(ctx context.Context, update *models.QuestionUpdate) error {
	key := s.getQuestionUpdateKey(update.FormID, update.QuestionID, update.UserID)
	data, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal question update: %w", err)
	}

	return s.client.Set(ctx, key, data, time.Hour).Err()
}

// Additional key generation methods

// getUserFormSessionKey generates key for user session in a specific form
func (s *Service) getUserFormSessionKey(userID, formID string) string {
	return fmt.Sprintf("%s:user_form_session:%s:%s", s.keyPrefix, userID, formID)
}

// getCursorPositionKey generates key for cursor position
func (s *Service) getCursorPositionKey(formID, userID string) string {
	return fmt.Sprintf("%s:cursor_position:%s:%s", s.keyPrefix, formID, userID)
}

// getQuestionUpdateKey generates key for question updates
func (s *Service) getQuestionUpdateKey(formID, questionID, userID string) string {
	return fmt.Sprintf("%s:question_update:%s:%s:%s:%d", s.keyPrefix, formID, questionID, userID, time.Now().Unix())
}
