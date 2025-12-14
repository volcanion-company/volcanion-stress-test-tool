-- Rollback: Remove performance indexes
-- Created: 2024-12-14

DROP INDEX IF EXISTS idx_test_runs_status;
DROP INDEX IF EXISTS idx_test_runs_completed_at;
DROP INDEX IF EXISTS idx_test_runs_plan_id;
DROP INDEX IF EXISTS idx_test_runs_status_completed;
DROP INDEX IF EXISTS idx_metrics_run_id;
DROP INDEX IF EXISTS idx_metrics_timestamp;
DROP INDEX IF EXISTS idx_metrics_run_timestamp;
DROP INDEX IF EXISTS idx_test_plans_name;
DROP INDEX IF EXISTS idx_test_plans_created_at;
DROP INDEX IF EXISTS idx_scenarios_name;
DROP INDEX IF EXISTS idx_scenario_executions_scenario_id;
DROP INDEX IF EXISTS idx_scenario_executions_status;
DROP INDEX IF EXISTS idx_api_keys_user_id;
DROP INDEX IF EXISTS idx_api_keys_key_hash;
DROP INDEX IF EXISTS idx_token_blacklist_expires_at;
