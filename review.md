# Code Review Report

## Findings (ordered by severity)
- Cancelling a run is overwritten as completed: when StopTest marks a run cancelled, monitorTestRun later sets it to completed, losing the cancelled state and end time source of truth (internal/domain/service/test_service.go).
- Unbounded request rate: generateRequests ticks every 1ms and fills the channel as fast as possible, so load is not a configured RPS and can overload the target unpredictably (internal/engine/scheduler.go).
- Latency retention is unbounded: each worker appends all latencies to a slice; long or high-RPS tests will spike memory/GC, and live P95/P99 are unavailable because percentiles are computed only at the end (internal/engine/worker.go, internal/engine/scheduler.go).
- HTTP client per worker: each worker creates its own http.Client/Transport, multiplying socket pools and file descriptors; not tied to MaxWorkers (internal/engine/worker.go).
- Config unused: MaxWorkers and DefaultTimeout are never applied; timeout falls back to a hard-coded 30s instead of config (internal/config/config.go, internal/domain/service/test_service.go).
- Prometheus collector unused: metrics are defined but never wired to the engine; gauges and counters stay empty (internal/metrics/collector.go, cmd/server/main.go).
- RPS fields never populated: CurrentRPS and RequestsPerSec stay zero during runs; TotalDurationMs set only after completion, so throughput is unknown live (internal/domain/model/metrics.go, internal/engine/load_generator.go).
- Handler error mapping: starting a test with a missing plan returns 500 instead of 404 because repository errors are not translated (internal/api/handler/test_run_handler.go, internal/domain/service/test_service.go).
- Graceful shutdown does not stop active tests: server shutdown does not signal/await generators; goroutines may outlive the process and runs remain “running” in memory (cmd/server/main.go).
- Metrics gauges never set: ActiveTests/ActiveWorkers gauges are defined but not updated (internal/metrics/collector.go).

## Recommended Fixes (implementation focused)
- Preserve cancellation: carry a stop reason or check run status before marking completed; ensure StopTest waits for scheduler exit and finalizes TotalDurationMs.
- Add explicit rate control: support target RPS (steady/step/ramp); derive ticker interval from desired rate instead of a fixed 1ms.
- Bound latency storage: replace slices with streaming histograms (HDR/t-digest) or ring buffers; compute rolling percentiles for live metrics.
- Share HTTP transport: use one tuned Transport (connection limits, keep-alive) across workers; enforce MaxWorkers and per-host limits.
- Apply config defaults: use DefaultTimeout when plan omits timeout_ms; cap or reject users beyond MaxWorkers.
- Wire Prometheus metrics: increment counters/histograms in worker execution; set gauges for active tests/workers with labels per run.
- Track live throughput: maintain per-interval counters to populate CurrentRPS and live RequestsPerSec.
- Map domain errors to HTTP codes: 404 for missing plan/run, 409 for already running.
- Graceful shutdown: stop all active tests, wait for workers, persist final metrics before exit.
- Update gauges: set ActiveTests/ActiveWorkers from the load generator lifecycle.

## Functional/Business Enhancements
- Persistence: add DB-backed repositories (PostgreSQL/Redis) for plans, runs, and metrics history.
- Scenario chaining: multi-step flows (login → token → subsequent calls) with variable extraction and templating.
- SLAs/assertions: allow thresholds (e.g., P95 < 300ms, error rate < 1%) and mark runs failed if breached.
- Rate models: fixed, step, ramp, spike/burst, arrivals-per-second (Poisson), per-endpoint weights.
- Data-driven tests: parameterize payloads/headers from CSV/JSON, random payload generators, think-time controls.
- Distributed workers: master/agent mode with gRPC, worker auto-registration, heartbeat, aggregated metrics.
- Reporting: export JSON/CSV summaries, compare multiple runs, generate shareable reports.
- Security/governance: API auth (API key/JWT), RBAC, audit logging.
- UX: Web UI or dashboard, WebSocket/SSE live metrics stream, CLI wrapper for scripting.
