-- Migration: Add performance indexes
-- Created: 2024-12-14

-- Index for test_runs queries by status
CREATE INDEX IF NOT EXISTS idx_test_runs_status ON test_runs(status);

-- Index for test_runs queries by completion time (for history/reports)
CREATE INDEX IF NOT EXISTS idx_test_runs_completed_at ON test_runs(completed_at DESC);

-- Index for test_runs by plan_id for filtering
CREATE INDEX IF NOT EXISTS idx_test_runs_plan_id ON test_runs(plan_id);

-- Composite index for common query pattern: status + completed_at
CREATE INDEX IF NOT EXISTS idx_test_runs_status_completed ON test_runs(status, completed_at DESC);

-- Index for metrics by run_id (critical for fetching test metrics)
CREATE INDEX IF NOT EXISTS idx_metrics_run_id ON metrics(run_id);

-- Index for metrics by timestamp for time-range queries
CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp DESC);

-- Composite index for metrics time-series queries
CREATE INDEX IF NOT EXISTS idx_metrics_run_timestamp ON metrics(run_id, timestamp DESC);

-- Index for test_plans by name for search
CREATE INDEX IF NOT EXISTS idx_test_plans_name ON test_plans(name);

-- Index for test_plans by creation time
CREATE INDEX IF NOT EXISTS idx_test_plans_created_at ON test_plans(created_at DESC);

-- Index for scenarios by name
CREATE INDEX IF NOT EXISTS idx_scenarios_name ON scenarios(name);

-- Index for scenario executions by scenario_id
CREATE INDEX IF NOT EXISTS idx_scenario_executions_scenario_id ON scenario_executions(scenario_id);

-- Index for scenario executions by status
CREATE INDEX IF NOT EXISTS idx_scenario_executions_status ON scenario_executions(status);

-- Index for API keys by user_id
CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);

-- Index for API keys by key_hash for fast lookup
CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);

-- Index for token blacklist by expiration for cleanup
CREATE INDEX IF NOT EXISTS idx_token_blacklist_expires_at ON token_blacklist(expires_at);

-- Analyze tables to update statistics after index creation
ANALYZE test_runs;
ANALYZE metrics;
ANALYZE test_plans;
