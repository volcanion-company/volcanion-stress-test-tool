package postgres

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/storage/repository"
)

type PostgresTestPlanRepository struct {
	db *sql.DB
}

func NewPostgresTestPlanRepository(db *sql.DB) *PostgresTestPlanRepository {
	return &PostgresTestPlanRepository{db: db}
}

func (r *PostgresTestPlanRepository) Create(plan *model.TestPlan) error {
	headers, err := json.Marshal(plan.Headers)
	if err != nil {
		return err
	}

	rateSteps, err := json.Marshal(plan.RateSteps)
	if err != nil {
		return err
	}

	slaConfig, err := json.Marshal(plan.SLA)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO test_plans (
			id, name, target_url, http_method, headers, body,
			concurrent_users, duration_seconds, target_rps, timeout_ms,
			rate_pattern, rate_steps, sla_config, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	now := time.Now()
	_, err = r.db.Exec(query,
		plan.ID, plan.Name, plan.TargetURL, plan.Method, headers, plan.Body,
		plan.Users, plan.DurationSec, plan.TargetRPS, plan.TimeoutMs,
		plan.RatePattern, rateSteps, slaConfig, now, now,
	)

	return err
}

func (r *PostgresTestPlanRepository) GetByID(id string) (*model.TestPlan, error) {
	query := `
		SELECT id, name, target_url, http_method, headers, body,
		       concurrent_users, duration_seconds, target_rps, timeout_ms,
		       rate_pattern, rate_steps, sla_config, created_at, updated_at
		FROM test_plans WHERE id = $1
	`

	plan := &model.TestPlan{}
	var headersJSON, rateStepsJSON, slaConfigJSON []byte
	var createdAt, updatedAt time.Time

	err := r.db.QueryRow(query, id).Scan(
		&plan.ID, &plan.Name, &plan.TargetURL, &plan.Method, &headersJSON, &plan.Body,
		&plan.Users, &plan.DurationSec, &plan.TargetRPS, &plan.TimeoutMs,
		&plan.RatePattern, &rateStepsJSON, &slaConfigJSON, &createdAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, repository.ErrTestPlanNotFound
	}
	if err != nil {
		return nil, err
	}

	if len(headersJSON) > 0 {
		if err := json.Unmarshal(headersJSON, &plan.Headers); err != nil {
			return nil, err
		}
	}

	if len(rateStepsJSON) > 0 {
		if err := json.Unmarshal(rateStepsJSON, &plan.RateSteps); err != nil {
			return nil, err
		}
	}

	if len(slaConfigJSON) > 0 {
		if err := json.Unmarshal(slaConfigJSON, &plan.SLA); err != nil {
			return nil, err
		}
	}

	return plan, nil
}

func (r *PostgresTestPlanRepository) GetAll() ([]*model.TestPlan, error) {
	query := `
		SELECT id, name, target_url, http_method, headers, body,
		       concurrent_users, duration_seconds, target_rps, timeout_ms,
		       rate_pattern, rate_steps, sla_config, created_at, updated_at
		FROM test_plans
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*model.TestPlan
	for rows.Next() {
		plan := &model.TestPlan{}
		var headersJSON, rateStepsJSON, slaConfigJSON []byte
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&plan.ID, &plan.Name, &plan.TargetURL, &plan.Method, &headersJSON, &plan.Body,
			&plan.Users, &plan.DurationSec, &plan.TargetRPS, &plan.TimeoutMs,
			&plan.RatePattern, &rateStepsJSON, &slaConfigJSON, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(headersJSON) > 0 {
			json.Unmarshal(headersJSON, &plan.Headers)
		}

		if len(rateStepsJSON) > 0 {
			json.Unmarshal(rateStepsJSON, &plan.RateSteps)
		}

		if len(slaConfigJSON) > 0 {
			json.Unmarshal(slaConfigJSON, &plan.SLA)
		}

		plans = append(plans, plan)
	}

	return plans, rows.Err()
}

func (r *PostgresTestPlanRepository) Delete(id string) error {
	query := `DELETE FROM test_plans WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrTestPlanNotFound
	}

	return nil
}
