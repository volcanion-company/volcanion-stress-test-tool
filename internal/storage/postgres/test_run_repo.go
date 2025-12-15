package postgres

import (
	"database/sql"
	"time"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/storage/repository"
)

//nolint:revive // exported name intentionally includes package name for clarity
type PostgresTestRunRepository struct {
	db *sql.DB
}

func NewPostgresTestRunRepository(db *sql.DB) *PostgresTestRunRepository {
	return &PostgresTestRunRepository{db: db}
}

func (r *PostgresTestRunRepository) Create(run *model.TestRun) error {
	query := `
		INSERT INTO test_runs (id, plan_id, status, stop_reason, started_at, completed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	now := time.Now()
	_, err := r.db.Exec(query,
		run.ID, run.PlanID, run.Status, stringOrNil(run.StopReason),
		timeOrNil(run.StartAt), timeOrNil(run.EndAt), now,
	)

	return err
}

func (r *PostgresTestRunRepository) GetByID(id string) (*model.TestRun, error) {
	query := `
		SELECT id, plan_id, status, stop_reason, started_at, completed_at, created_at
		FROM test_runs WHERE id = $1
	`

	run := &model.TestRun{}
	var stopReason sql.NullString
	var startedAt, completedAt sql.NullTime
	var createdAt time.Time

	err := r.db.QueryRow(query, id).Scan(
		&run.ID, &run.PlanID, &run.Status, &stopReason,
		&startedAt, &completedAt, &createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, repository.ErrTestRunNotFound
	}
	if err != nil {
		return nil, err
	}

	if stopReason.Valid {
		r := model.StopReason(stopReason.String)
		run.StopReason = &r
	}
	if startedAt.Valid {
		run.StartAt = startedAt.Time
	}
	if completedAt.Valid {
		run.EndAt = &completedAt.Time
	}

	return run, nil
}

func (r *PostgresTestRunRepository) GetAll() ([]*model.TestRun, error) {
	query := `
		SELECT id, plan_id, status, stop_reason, started_at, completed_at, created_at
		FROM test_runs
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []*model.TestRun
	for rows.Next() {
		run := &model.TestRun{}
		var stopReason sql.NullString
		var startedAt, completedAt sql.NullTime
		var createdAt time.Time

		err := rows.Scan(
			&run.ID, &run.PlanID, &run.Status, &stopReason,
			&startedAt, &completedAt, &createdAt,
		)
		if err != nil {
			return nil, err
		}

		if stopReason.Valid {
			r := model.StopReason(stopReason.String)
			run.StopReason = &r
		}
		if startedAt.Valid {
			run.StartAt = startedAt.Time
		}
		if completedAt.Valid {
			run.EndAt = &completedAt.Time
		}

		runs = append(runs, run)
	}

	return runs, rows.Err()
}

func (r *PostgresTestRunRepository) Update(run *model.TestRun) error {
	query := `
		UPDATE test_runs
		SET status = $2, stop_reason = $3, started_at = $4, completed_at = $5
		WHERE id = $1
	`

	result, err := r.db.Exec(query,
		run.ID, run.Status, stringOrNil(run.StopReason),
		timeOrNil(run.StartAt), timeOrNil(run.EndAt),
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrTestRunNotFound
	}

	return nil
}

func (r *PostgresTestRunRepository) Delete(id string) error {
	query := `DELETE FROM test_runs WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrTestRunNotFound
	}

	return nil
}

// Helper functions
func stringOrNil(s *model.StopReason) interface{} {
	if s == nil {
		return nil
	}
	return string(*s)
}

func timeOrNil(t interface{}) interface{} {
	switch v := t.(type) {
	case time.Time:
		if v.IsZero() {
			return nil
		}
		return v
	case *time.Time:
		if v == nil {
			return nil
		}
		return v
	default:
		return nil
	}
}
