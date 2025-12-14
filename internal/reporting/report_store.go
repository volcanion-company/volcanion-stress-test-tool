package reporting

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

// ReportStore manages shareable reports
type ReportStore struct {
	mu      sync.RWMutex
	reports map[string]*StoredReport
}

// StoredReport represents a stored report
type StoredReport struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Format      ExportFormat      `json:"format"`
	Content     []byte            `json:"-"` // Raw content (not exposed in JSON)
	ContentType string            `json:"content_type"`
	CreatedAt   time.Time         `json:"created_at"`
	ExpiresAt   time.Time         `json:"expires_at"`
	AccessCount int               `json:"access_count"`
	Metadata    map[string]string `json:"metadata"`
}

// NewReportStore creates a new report store
func NewReportStore() *ReportStore {
	store := &ReportStore{
		reports: make(map[string]*StoredReport),
	}

	// Start cleanup goroutine
	go store.cleanupExpired()

	return store
}

// Store saves a report and returns a shareable ID
func (s *ReportStore) Store(title string, format ExportFormat, content []byte, ttl time.Duration, metadata map[string]string) (*StoredReport, error) {
	id, err := generateReportID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate report ID: %w", err)
	}

	contentType := s.getContentType(format)

	report := &StoredReport{
		ID:          id,
		Title:       title,
		Format:      format,
		Content:     content,
		ContentType: contentType,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(ttl),
		AccessCount: 0,
		Metadata:    metadata,
	}

	s.mu.Lock()
	s.reports[id] = report
	s.mu.Unlock()

	return report, nil
}

// Get retrieves a report by ID and increments access count
func (s *ReportStore) Get(id string) (*StoredReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	report, exists := s.reports[id]
	if !exists {
		return nil, fmt.Errorf("report not found")
	}

	if time.Now().After(report.ExpiresAt) {
		delete(s.reports, id)
		return nil, fmt.Errorf("report has expired")
	}

	report.AccessCount++
	return report, nil
}

// GetInfo retrieves report metadata without incrementing access count
func (s *ReportStore) GetInfo(id string) (*StoredReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	report, exists := s.reports[id]
	if !exists {
		return nil, fmt.Errorf("report not found")
	}

	if time.Now().After(report.ExpiresAt) {
		return nil, fmt.Errorf("report has expired")
	}

	// Return copy without content
	info := &StoredReport{
		ID:          report.ID,
		Title:       report.Title,
		Format:      report.Format,
		ContentType: report.ContentType,
		CreatedAt:   report.CreatedAt,
		ExpiresAt:   report.ExpiresAt,
		AccessCount: report.AccessCount,
		Metadata:    report.Metadata,
	}

	return info, nil
}

// Delete removes a report by ID
func (s *ReportStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.reports[id]; !exists {
		return fmt.Errorf("report not found")
	}

	delete(s.reports, id)
	return nil
}

// List returns all active (non-expired) reports
func (s *ReportStore) List() []*StoredReport {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	reports := make([]*StoredReport, 0)

	for _, report := range s.reports {
		if now.Before(report.ExpiresAt) {
			// Return copy without content
			info := &StoredReport{
				ID:          report.ID,
				Title:       report.Title,
				Format:      report.Format,
				ContentType: report.ContentType,
				CreatedAt:   report.CreatedAt,
				ExpiresAt:   report.ExpiresAt,
				AccessCount: report.AccessCount,
				Metadata:    report.Metadata,
			}
			reports = append(reports, info)
		}
	}

	return reports
}

// ExtendExpiration extends the expiration time of a report
func (s *ReportStore) ExtendExpiration(id string, additionalTTL time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	report, exists := s.reports[id]
	if !exists {
		return fmt.Errorf("report not found")
	}

	report.ExpiresAt = report.ExpiresAt.Add(additionalTTL)
	return nil
}

// cleanupExpired periodically removes expired reports
func (s *ReportStore) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for id, report := range s.reports {
			if now.After(report.ExpiresAt) {
				delete(s.reports, id)
			}
		}
		s.mu.Unlock()
	}
}

// getContentType returns the appropriate content type for a format
func (s *ReportStore) getContentType(format ExportFormat) string {
	switch format {
	case FormatJSON:
		return "application/json"
	case FormatCSV:
		return "text/csv"
	case FormatHTML:
		return "text/html"
	default:
		return "application/octet-stream"
	}
}

// GetStats returns statistics about stored reports
func (s *ReportStore) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalReports := len(s.reports)
	expiredCount := 0
	totalAccesses := 0
	now := time.Now()

	formatCounts := make(map[ExportFormat]int)

	for _, report := range s.reports {
		if now.After(report.ExpiresAt) {
			expiredCount++
		}
		totalAccesses += report.AccessCount
		formatCounts[report.Format]++
	}

	return map[string]interface{}{
		"total_reports":    totalReports,
		"active_reports":   totalReports - expiredCount,
		"expired_reports":  expiredCount,
		"total_accesses":   totalAccesses,
		"format_breakdown": formatCounts,
	}
}

// generateReportID generates a random URL-safe report ID
func generateReportID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
