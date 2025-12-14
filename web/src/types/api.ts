export interface TestPlan {
  id: string
  name: string
  target_url: string
  method: string
  headers: Record<string, string>
  body: string
  concurrent_users: number
  ramp_up_sec: number
  duration_sec: number
  timeout_ms: number
  rate_pattern: 'fixed' | 'step' | 'ramp' | 'spike'
  target_rps?: number
  rate_steps?: RateStep[]
  sla_config?: SLAConfig
  created_at: string
}

export interface RateStep {
  duration_sec: number
  rps: number
}

export interface SLAConfig {
  max_p95_latency_ms?: number
  max_p99_latency_ms?: number
  max_avg_latency_ms?: number
  min_success_rate?: number
  max_error_rate?: number
  min_rps?: number
}

export interface TestRun {
  id: string
  plan_id: string
  status: 'running' | 'completed' | 'failed' | 'cancelled'
  stop_reason?: 'completed' | 'cancelled' | 'failed'
  start_at: string
  end_at?: string
}

export interface Metrics {
  run_id: string
  total_requests: number
  success_requests: number
  failed_requests: number
  avg_latency_ms: number
  min_latency_ms: number
  max_latency_ms: number
  p50_latency_ms: number
  p75_latency_ms: number
  p95_latency_ms: number
  p99_latency_ms: number
  requests_per_sec: number
  current_rps: number
  active_workers: number
  total_duration_ms: number
  status_codes: Record<string, number>
  errors: Record<string, number>
}

// Alias for WebSocket compatibility
export type TestMetrics = Metrics

export interface CreateTestPlanRequest {
  name: string
  target_url: string
  method: string
  headers?: Record<string, string>
  body?: string
  concurrent_users: number
  ramp_up_sec?: number
  duration_sec: number
  timeout_ms?: number
  rate_pattern?: string
  target_rps?: number
  rate_steps?: RateStep[]
  sla_config?: SLAConfig
}
