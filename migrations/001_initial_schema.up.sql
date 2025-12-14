-- Create test_plans table
CREATE TABLE IF NOT EXISTS test_plans (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    target_url TEXT NOT NULL,
    http_method VARCHAR(10) NOT NULL DEFAULT 'GET',
    headers JSONB,
    body TEXT,
    concurrent_users INT NOT NULL DEFAULT 1,
    duration_seconds INT NOT NULL DEFAULT 60,
    target_rps INT DEFAULT 0,
    timeout_ms INT NOT NULL DEFAULT 30000,
    rate_pattern VARCHAR(20) DEFAULT 'fixed',
    rate_steps JSONB,
    sla_config JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_test_plans_name ON test_plans(name);
CREATE INDEX idx_test_plans_created_at ON test_plans(created_at DESC);

-- Create test_runs table
CREATE TABLE IF NOT EXISTS test_runs (
    id VARCHAR(36) PRIMARY KEY,
    plan_id VARCHAR(36) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    stop_reason VARCHAR(20),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (plan_id) REFERENCES test_plans(id) ON DELETE CASCADE
);

CREATE INDEX idx_test_runs_plan_id ON test_runs(plan_id);
CREATE INDEX idx_test_runs_status ON test_runs(status);
CREATE INDEX idx_test_runs_created_at ON test_runs(created_at DESC);
CREATE INDEX idx_test_runs_started_at ON test_runs(started_at DESC);

-- Create metrics_snapshots table for time-series data
CREATE TABLE IF NOT EXISTS metrics_snapshots (
    id BIGSERIAL PRIMARY KEY,
    run_id VARCHAR(36) NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    total_requests BIGINT NOT NULL DEFAULT 0,
    successful_requests BIGINT NOT NULL DEFAULT 0,
    failed_requests BIGINT NOT NULL DEFAULT 0,
    total_duration_ms BIGINT NOT NULL DEFAULT 0,
    avg_response_time_ms FLOAT NOT NULL DEFAULT 0,
    min_response_time_ms FLOAT NOT NULL DEFAULT 0,
    max_response_time_ms FLOAT NOT NULL DEFAULT 0,
    p50_ms FLOAT NOT NULL DEFAULT 0,
    p95_ms FLOAT NOT NULL DEFAULT 0,
    p99_ms FLOAT NOT NULL DEFAULT 0,
    current_rps FLOAT NOT NULL DEFAULT 0,
    error_rate FLOAT NOT NULL DEFAULT 0,
    status_codes JSONB,
    errors JSONB,
    FOREIGN KEY (run_id) REFERENCES test_runs(id) ON DELETE CASCADE
);

CREATE INDEX idx_metrics_snapshots_run_id ON metrics_snapshots(run_id);
CREATE INDEX idx_metrics_snapshots_timestamp ON metrics_snapshots(run_id, timestamp DESC);

-- Create final_metrics table for aggregated results
CREATE TABLE IF NOT EXISTS final_metrics (
    run_id VARCHAR(36) PRIMARY KEY,
    total_requests BIGINT NOT NULL DEFAULT 0,
    successful_requests BIGINT NOT NULL DEFAULT 0,
    failed_requests BIGINT NOT NULL DEFAULT 0,
    total_duration_ms BIGINT NOT NULL DEFAULT 0,
    avg_response_time_ms FLOAT NOT NULL DEFAULT 0,
    min_response_time_ms FLOAT NOT NULL DEFAULT 0,
    max_response_time_ms FLOAT NOT NULL DEFAULT 0,
    p50_ms FLOAT NOT NULL DEFAULT 0,
    p95_ms FLOAT NOT NULL DEFAULT 0,
    p99_ms FLOAT NOT NULL DEFAULT 0,
    requests_per_sec FLOAT NOT NULL DEFAULT 0,
    error_rate FLOAT NOT NULL DEFAULT 0,
    status_codes JSONB,
    errors JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (run_id) REFERENCES test_runs(id) ON DELETE CASCADE
);

CREATE INDEX idx_final_metrics_created_at ON final_metrics(created_at DESC);
