package postgres

import (
	"database/sql"
	"encoding/json"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
)

//nolint:revive // exported name intentionally includes package name for clarity
type PostgresMetricsRepository struct {
	db *sql.DB
}

func NewPostgresMetricsRepository(db *sql.DB) *PostgresMetricsRepository {
	return &PostgresMetricsRepository{db: db}
}

func (r *PostgresMetricsRepository) Save(metrics *model.Metrics) error {
	statusCodes, err := json.Marshal(metrics.StatusCodes)
	if err != nil {
		return err
	}

	errors, err := json.Marshal(metrics.Errors)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO final_metrics (
			run_id, total_requests, successful_requests, failed_requests,
			total_duration_ms, avg_response_time_ms, min_response_time_ms, max_response_time_ms,
			p50_ms, p95_ms, p99_ms, requests_per_sec, error_rate,
			status_codes, errors
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (run_id) DO UPDATE SET
			total_requests = EXCLUDED.total_requests,
			successful_requests = EXCLUDED.successful_requests,
			failed_requests = EXCLUDED.failed_requests,
			total_duration_ms = EXCLUDED.total_duration_ms,
			avg_response_time_ms = EXCLUDED.avg_response_time_ms,
			min_response_time_ms = EXCLUDED.min_response_time_ms,
			max_response_time_ms = EXCLUDED.max_response_time_ms,
			p50_ms = EXCLUDED.p50_ms,
			p95_ms = EXCLUDED.p95_ms,
			p99_ms = EXCLUDED.p99_ms,
			requests_per_sec = EXCLUDED.requests_per_sec,
			error_rate = EXCLUDED.error_rate,
			status_codes = EXCLUDED.status_codes,
			errors = EXCLUDED.errors
	`

	// Calculate error rate
	errorRate := 0.0
	if metrics.TotalRequests > 0 {
		errorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests) * 100
	}

	_, err = r.db.Exec(query,
		metrics.RunID, metrics.TotalRequests, metrics.SuccessRequests, metrics.FailedRequests,
		metrics.TotalDurationMs, metrics.AvgLatencyMs, metrics.MinLatencyMs, metrics.MaxLatencyMs,
		metrics.P50LatencyMs, metrics.P95LatencyMs, metrics.P99LatencyMs, metrics.RequestsPerSec, errorRate,
		statusCodes, errors,
	)

	return err
}

func (r *PostgresMetricsRepository) GetByRunID(runID string) (*model.Metrics, error) {
	query := `
		SELECT run_id, total_requests, successful_requests, failed_requests,
		       total_duration_ms, avg_response_time_ms, min_response_time_ms, max_response_time_ms,
		       p50_ms, p95_ms, p99_ms, requests_per_sec, error_rate,
		       status_codes, errors
		FROM final_metrics WHERE run_id = $1
	`

	metrics := &model.Metrics{}
	var statusCodesJSON, errorsJSON []byte

	var errorRate float64
	err := r.db.QueryRow(query, runID).Scan(
		&metrics.RunID, &metrics.TotalRequests, &metrics.SuccessRequests, &metrics.FailedRequests,
		&metrics.TotalDurationMs, &metrics.AvgLatencyMs, &metrics.MinLatencyMs, &metrics.MaxLatencyMs,
		&metrics.P50LatencyMs, &metrics.P95LatencyMs, &metrics.P99LatencyMs, &metrics.RequestsPerSec, &errorRate,
		&statusCodesJSON, &errorsJSON,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(statusCodesJSON) > 0 {
		if err := json.Unmarshal(statusCodesJSON, &metrics.StatusCodes); err != nil {
			return nil, err
		}
	}

	if len(errorsJSON) > 0 {
		if err := json.Unmarshal(errorsJSON, &metrics.Errors); err != nil {
			return nil, err
		}
	}

	return metrics, nil
}

func (r *PostgresMetricsRepository) Delete(runID string) error {
	query := `DELETE FROM final_metrics WHERE run_id = $1`
	_, err := r.db.Exec(query, runID)
	return err
}
