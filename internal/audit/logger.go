package audit

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Logger handles audit logging
type Logger struct {
	events   []AuditEvent
	mu       sync.RWMutex
	zapLog   *zap.Logger
	maxSize  int // Maximum events to keep in memory
	filePath string
}

// NewLogger creates a new audit logger
func NewLogger(zapLog *zap.Logger, maxSize int) *Logger {
	return &Logger{
		events:  make([]AuditEvent, 0),
		zapLog:  zapLog,
		maxSize: maxSize,
	}
}

// Log records an audit event
func (l *Logger) Log(event AuditEvent) {
	event.ID = uuid.New().String()
	event.Timestamp = time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Add to in-memory store
	l.events = append(l.events, event)

	// Trim if exceeds max size
	if len(l.events) > l.maxSize {
		l.events = l.events[len(l.events)-l.maxSize:]
	}

	// Also log to zap logger for structured logging
	l.zapLog.Info("audit_event",
		zap.String("event_id", event.ID),
		zap.String("event_type", string(event.EventType)),
		zap.String("user_id", event.UserID),
		zap.String("username", event.Username),
		zap.String("ip", event.IPAddress),
		zap.String("method", event.Method),
		zap.String("path", event.Path),
		zap.Int("status", event.StatusCode),
		zap.Int64("duration_ms", event.Duration),
		zap.String("resource_id", event.ResourceID),
		zap.Any("details", event.Details),
	)
}

// Query retrieves audit events based on filter criteria
func (l *Logger) Query(filter AuditFilter) []AuditEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	filtered := make([]AuditEvent, 0)

	for _, event := range l.events {
		// Apply filters
		if filter.StartTime != nil && event.Timestamp.Before(*filter.StartTime) {
			continue
		}
		if filter.EndTime != nil && event.Timestamp.After(*filter.EndTime) {
			continue
		}
		if filter.UserID != "" && event.UserID != filter.UserID {
			continue
		}
		if filter.EventType != "" && event.EventType != filter.EventType {
			continue
		}
		if filter.IPAddress != "" && event.IPAddress != filter.IPAddress {
			continue
		}
		if filter.ResourceID != "" && event.ResourceID != filter.ResourceID {
			continue
		}

		filtered = append(filtered, event)
	}

	// Apply pagination
	start := filter.Offset
	if start > len(filtered) {
		return []AuditEvent{}
	}

	end := start + filter.Limit
	if filter.Limit == 0 || end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end]
}

// GetAll returns all audit events
func (l *Logger) GetAll() []AuditEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	events := make([]AuditEvent, len(l.events))
	copy(events, l.events)
	return events
}

// Export exports audit logs as JSON
func (l *Logger) Export() ([]byte, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return json.MarshalIndent(l.events, "", "  ")
}

// ExportFiltered exports filtered audit logs as JSON
func (l *Logger) ExportFiltered(filter AuditFilter) ([]byte, error) {
	filtered := l.Query(filter)
	return json.MarshalIndent(filtered, "", "  ")
}

// Count returns the total number of audit events
func (l *Logger) Count() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.events)
}

// Clear removes all audit events (use with caution)
func (l *Logger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = make([]AuditEvent, 0)
}
