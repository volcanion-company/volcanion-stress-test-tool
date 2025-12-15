package postgres

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
)

// MetricsSnapshot represents a point-in-time snapshot of metrics
type MetricsSnapshot struct {
	ID                 int64
	RunID              string
	Timestamp          time.Time
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	TotalDurationMs    int64
	AvgResponseTimeMs  float64
	MinResponseTimeMs  float64
	MaxResponseTimeMs  float64
	P50Ms              float64
	P95Ms              float64
	P99Ms              float64
	CurrentRPS         float64
	ErrorRate          float64
	StatusCodes        map[string]int
	Errors             map[string]int
}

// MetricsSnapshotRepository handles time-series metrics storage
type MetricsSnapshotRepository struct {
	db *sql.DB
}

func NewMetricsSnapshotRepository(db *sql.DB) *MetricsSnapshotRepository {
	return &MetricsSnapshotRepository{db: db}
}

// SaveSnapshot stores a metrics snapshot at current time
func (r *MetricsSnapshotRepository) SaveSnapshot(runID string, metrics *model.Metrics) error {
	statusCodes, err := json.Marshal(metrics.StatusCodes)
	if err != nil {
		return err
	}

	errors, err := json.Marshal(metrics.Errors)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO metrics_snapshots (
			run_id, timestamp, total_requests, successful_requests, failed_requests,
			total_duration_ms, avg_response_time_ms, min_response_time_ms, max_response_time_ms,
			p50_ms, p95_ms, p99_ms, current_rps, error_rate,
			status_codes, errors
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	// Calculate error rate
	errorRate := 0.0
	if metrics.TotalRequests > 0 {
		errorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests) * 100
	}

	_, err = r.db.Exec(query,
		runID, time.Now(),
		metrics.TotalRequests, metrics.SuccessRequests, metrics.FailedRequests,
		metrics.TotalDurationMs, metrics.AvgLatencyMs, metrics.MinLatencyMs, metrics.MaxLatencyMs,
		metrics.P50LatencyMs, metrics.P95LatencyMs, metrics.P99LatencyMs, metrics.CurrentRPS, errorRate,
		statusCodes, errors,
	)

	return err
}

// GetSnapshots retrieves all snapshots for a run within a time range
func (r *MetricsSnapshotRepository) GetSnapshots(runID string, start, end time.Time) ([]MetricsSnapshot, error) {
	query := `
		SELECT id, run_id, timestamp, total_requests, successful_requests, failed_requests,
		       total_duration_ms, avg_response_time_ms, min_response_time_ms, max_response_time_ms,
		       p50_ms, p95_ms, p99_ms, current_rps, error_rate,
		       status_codes, errors
		FROM metrics_snapshots
		WHERE run_id = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp ASC
	`

	rows, err := r.db.Query(query, runID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []MetricsSnapshot
	for rows.Next() {
		var snapshot MetricsSnapshot
		var statusCodesJSON, errorsJSON []byte

		err := rows.Scan(
			&snapshot.ID, &snapshot.RunID, &snapshot.Timestamp,
			&snapshot.TotalRequests, &snapshot.SuccessfulRequests, &snapshot.FailedRequests,
			&snapshot.TotalDurationMs, &snapshot.AvgResponseTimeMs, &snapshot.MinResponseTimeMs, &snapshot.MaxResponseTimeMs,
			&snapshot.P50Ms, &snapshot.P95Ms, &snapshot.P99Ms, &snapshot.CurrentRPS, &snapshot.ErrorRate,
			&statusCodesJSON, &errorsJSON,
		)
		if err != nil {
			return nil, err
		}

		if len(statusCodesJSON) > 0 {
			if err := json.Unmarshal(statusCodesJSON, &snapshot.StatusCodes); err != nil {
				return nil, err
			}
		}

		if len(errorsJSON) > 0 {
			if err := json.Unmarshal(errorsJSON, &snapshot.Errors); err != nil {
				return nil, err
			}
		}

		snapshots = append(snapshots, snapshot)
	}

	return snapshots, rows.Err()
}

// GetLatestSnapshots retrieves the N most recent snapshots for a run
func (r *MetricsSnapshotRepository) GetLatestSnapshots(runID string, limit int) ([]MetricsSnapshot, error) {
	query := `
		SELECT id, run_id, timestamp, total_requests, successful_requests, failed_requests,
		       total_duration_ms, avg_response_time_ms, min_response_time_ms, max_response_time_ms,
		       p50_ms, p95_ms, p99_ms, current_rps, error_rate,
		       status_codes, errors
		FROM metrics_snapshots
		WHERE run_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, runID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []MetricsSnapshot
	for rows.Next() {
		var snapshot MetricsSnapshot
		var statusCodesJSON, errorsJSON []byte

		err := rows.Scan(
			&snapshot.ID, &snapshot.RunID, &snapshot.Timestamp,
			&snapshot.TotalRequests, &snapshot.SuccessfulRequests, &snapshot.FailedRequests,
			&snapshot.TotalDurationMs, &snapshot.AvgResponseTimeMs, &snapshot.MinResponseTimeMs, &snapshot.MaxResponseTimeMs,
			&snapshot.P50Ms, &snapshot.P95Ms, &snapshot.P99Ms, &snapshot.CurrentRPS, &snapshot.ErrorRate,
			&statusCodesJSON, &errorsJSON,
		)
		if err != nil {
			return nil, err
		}

		if len(statusCodesJSON) > 0 {
			if err := json.Unmarshal(statusCodesJSON, &snapshot.StatusCodes); err != nil {
				return nil, err
			}
		}

		if len(errorsJSON) > 0 {
			if err := json.Unmarshal(errorsJSON, &snapshot.Errors); err != nil {
				return nil, err
			}
		}

		snapshots = append(snapshots, snapshot)
	}

	return snapshots, rows.Err()
}

// DeleteOldSnapshots removes snapshots older than the retention period
func (r *MetricsSnapshotRepository) DeleteOldSnapshots(retentionDays int) error {
	query := `DELETE FROM metrics_snapshots WHERE timestamp < $1`
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	_, err := r.db.Exec(query, cutoff)
	return err
}

// DeleteSnapshotsByRunID removes all snapshots for a specific run
func (r *MetricsSnapshotRepository) DeleteSnapshotsByRunID(runID string) error {
	query := `DELETE FROM metrics_snapshots WHERE run_id = $1`
	_, err := r.db.Exec(query, runID)
	return err
}
