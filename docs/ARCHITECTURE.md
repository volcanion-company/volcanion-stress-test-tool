# Architecture

This document describes the architecture of Volcanion Stress Test Tool.

## Overview

Volcanion is a distributed HTTP load testing tool with the following components:

- **Backend API Server** - Go/Gin REST API with WebSocket support
- **Load Test Engine** - High-performance concurrent request generator
- **Web Frontend** - React SPA with real-time monitoring
- **CLI Tool** - Command-line interface for automation
- **Database** - PostgreSQL for persistence
- **Observability Stack** - Prometheus, Grafana, Jaeger

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              Clients                                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │
│  │   Web UI    │  │    CLI      │  │  CI/CD      │  │  External   │     │
│  │  (React)    │  │  (Cobra)    │  │  Pipeline   │  │    API      │     │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘     │
└─────────┼────────────────┼────────────────┼────────────────┼────────────┘
          │                │                │                │
          │         HTTP/WebSocket          │                │
          └────────────────┼────────────────┘────────────────┘
                           │
┌──────────────────────────┼──────────────────────────────────────────────┐
│                          │           NGINX (Optional)                   │
│                    ┌─────┴─────┐                                        │
│                    │  Reverse  │    - SSL Termination                   │
│                    │   Proxy   │    - Load Balancing                    │
│                    └─────┬─────┘    - Static Files                      │
└──────────────────────────┼──────────────────────────────────────────────┘
                           │
┌──────────────────────────┼──────────────────────────────────────────────┐
│                    API Server (Gin)                                     │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                     Middleware Stack                              │  │
│  │  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐   │  │
│  │  │ Req  │→│ Recov│→│ Log  │→│Metric│→│Trace │→│ CORS │→│ Auth │   │  │
│  │  │  ID  │ │ ery  │ │ ging │ │  s   │ │ ing  │ │      │ │      │   │  │
│  │  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘ └──────┘ └──────┘   │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                        API Handlers                               │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌─────────┐  │  │
│  │  │   Auth   │ │TestPlan  │ │ TestRun  │ │ Scenario │ │ Report  │  │  │
│  │  │ Handler  │ │ Handler  │ │ Handler  │ │ Handler  │ │ Handler │  │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └─────────┘  │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐                           │  │
│  │  │WebSocket │ │  Audit   │ │ Metrics  │                           │  │
│  │  │ Handler  │ │ Handler  │ │ Handler  │                           │  │
│  │  └──────────┘ └──────────┘ └──────────┘                           │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                       Domain Services                             │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐              │  │
│  │  │  Test    │ │ Scenario │ │  Report  │ │   Auth   │              │  │
│  │  │ Service  │ │ Service  │ │ Service  │ │ Service  │              │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘              │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                     Load Test Engine                              │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐              │  │
│  │  │Scheduler │ │ Workers  │ │ Metrics  │ │ Template │              │  │
│  │  │          │ │  Pool    │ │Collector │ │  Engine  │              │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘              │  │
│  └───────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
          │                    │                    │
          ▼                    ▼                    ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│   PostgreSQL    │  │   Prometheus    │  │     Jaeger      │
│   (Persistence) │  │   (Metrics)     │  │   (Tracing)     │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

---

## Component Details

### API Server

The API server is built with [Gin](https://github.com/gin-gonic/gin) web framework.

#### Middleware Stack

| Order | Middleware | Purpose |
|-------|------------|---------|
| 1 | RequestID | Generate unique request ID for tracing |
| 2 | Recovery | Catch panics and return 500 errors |
| 3 | Logging | Structured request/response logging |
| 4 | Metrics | Prometheus metrics collection |
| 5 | Tracing | OpenTelemetry distributed tracing |
| 6 | CORS | Cross-Origin Resource Sharing |
| 7 | RateLimit | Token bucket rate limiting |
| 8 | Auth | JWT/API key authentication |
| 9 | Audit | Audit logging for sensitive operations |

#### Handlers

| Handler | Responsibility |
|---------|----------------|
| AuthHandler | Login, JWT tokens, API key management |
| TestPlanHandler | CRUD for test plans |
| TestRunHandler | Start/stop tests, get results |
| ScenarioHandler | Multi-step test scenarios |
| ReportHandler | Report generation and export |
| WebSocketHandler | Real-time metrics streaming |
| AuditHandler | Audit log queries |

---

### Load Test Engine

The engine is responsible for generating HTTP load.

```
┌─────────────────────────────────────────────────────────────────┐
│                        Scheduler                                │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    Test Plan                            │    │
│  │  - Target URL      - Concurrent Users                   │    │
│  │  - HTTP Method     - Duration                           │    │
│  │  - Headers/Body    - Ramp-up Time                       │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                  │
│                              ▼                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                   Worker Pool                           │    │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐       ┌─────────┐  │    │
│  │  │Worker 1 │ │Worker 2 │ │Worker 3 │  ...  │Worker N │  │    │
│  │  └────┬────┘ └────┬────┘ └────┬────┘       └────┬────┘  │    │
│  │       │           │           │                 │       │    │
│  │       └───────────┴───────────┴─────────────────┘       │    │
│  │                           │                             │    │
│  │                    Shared HTTP Client                   │    │
│  │              (Connection Pool, Keep-Alive)              │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                  │
│                              ▼                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                  Metrics Collector                      │    │
│  │  - Request count    - Latency (p50, p95, p99)           │    │
│  │  - Success/Failure  - Throughput (RPS)                  │    │
│  │  - Status codes     - Active workers                    │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

#### Scheduler

- Manages test lifecycle (start, stop, pause)
- Controls worker pool size based on ramp-up configuration
- Coordinates graceful shutdown

#### Workers

- Each worker runs in a separate goroutine
- Executes HTTP requests against target
- Reports results to metrics collector
- Supports think time between requests

#### HTTP Client

- Shared `http.Transport` with connection pooling
- Configurable timeouts and keep-alive
- TLS configuration for HTTPS targets

#### Template Engine

Supports dynamic request content:
- `{{variable}}` - Variable substitution
- `{{$random}}` - Random string
- `{{$uuid}}` - Random UUID
- `{{$timestamp}}` - Current timestamp

---

### Web Frontend

React 18 SPA with modern tooling.

```
┌─────────────────────────────────────────────────────────────────┐
│                        React App                                │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    App Router                           │    │
│  │  / → Dashboard                                          │    │
│  │  /test-plans → TestPlanList                             │    │
│  │  /test-plans/:id → TestPlanDetail                       │    │
│  │  /test-runs → TestRunList                               │    │
│  │  /test-runs/:id → TestRunDetail (Real-time)             │    │
│  │  /scenarios → ScenarioList                              │    │
│  │  /reports → ReportList                                  │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    State Management                     │    │
│  │  ┌───────────────┐  ┌───────────────┐                   │    │
│  │  │ TanStack Query │  │  AuthContext  │                  │    │
│  │  │ (Server State) │  │ (Client State)│                  │    │
│  │  └───────────────┘  └───────────────┘                   │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    Custom Hooks                         │    │
│  │  useWebSocket │ useApiError │ useNetworkStatus          │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    UI Components                        │    │
│  │  Button │ Card │ Modal │ Table │ Chart │ Form           │    │
│  │  ErrorBoundary │ LoadingSpinner │ Skeleton              │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

#### Technology Stack

| Technology | Purpose |
|------------|---------|
| React 18 | UI framework |
| TypeScript | Type safety |
| Vite | Build tool |
| TanStack Query | Server state management |
| React Router | Client-side routing |
| Tailwind CSS | Styling |
| Recharts | Data visualization |
| Lucide Icons | Icon library |

---

### Database Schema

```sql
-- Users (managed externally or in-memory for demo)
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Test Plans
CREATE TABLE test_plans (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    target_url VARCHAR(2048) NOT NULL,
    method VARCHAR(10) NOT NULL,
    headers JSONB DEFAULT '{}',
    body TEXT,
    concurrent_users INT NOT NULL DEFAULT 10,
    duration_seconds INT NOT NULL DEFAULT 60,
    ramp_up_seconds INT DEFAULT 0,
    think_time_ms INT DEFAULT 0,
    timeout_ms INT DEFAULT 30000,
    variables JSONB DEFAULT '{}',
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Test Runs
CREATE TABLE test_runs (
    id UUID PRIMARY KEY,
    plan_id UUID REFERENCES test_plans(id),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    stop_reason VARCHAR(50),
    start_at TIMESTAMP,
    end_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Test Metrics (aggregated results)
CREATE TABLE test_metrics (
    id UUID PRIMARY KEY,
    run_id UUID REFERENCES test_runs(id),
    total_requests BIGINT DEFAULT 0,
    successful_requests BIGINT DEFAULT 0,
    failed_requests BIGINT DEFAULT 0,
    avg_latency_ms FLOAT,
    min_latency_ms FLOAT,
    max_latency_ms FLOAT,
    p50_latency_ms FLOAT,
    p95_latency_ms FLOAT,
    p99_latency_ms FLOAT,
    requests_per_second FLOAT,
    status_codes JSONB DEFAULT '{}',
    errors JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Scenarios
CREATE TABLE scenarios (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    steps JSONB NOT NULL,
    variables JSONB DEFAULT '{}',
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Audit Logs
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    user_id UUID,
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100),
    resource_id UUID,
    details JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_test_runs_plan_id ON test_runs(plan_id);
CREATE INDEX idx_test_runs_status ON test_runs(status);
CREATE INDEX idx_test_runs_start_at ON test_runs(start_at DESC);
CREATE INDEX idx_test_metrics_run_id ON test_metrics(run_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
```

---

### Security Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Security Layers                             │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                  Transport Security                     │    │
│  │  - HTTPS/TLS termination at nginx                       │    │
│  │  - Secure WebSocket (wss://)                            │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                  Authentication                         │    │
│  │  - JWT tokens (HS256, 24h expiry)                       │    │
│  │  - API keys (for automation)                            │    │
│  │  - bcrypt password hashing                              │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                   Authorization                         │    │
│  │  - Role-based access control (admin, user, readonly)    │    │
│  │  - Resource-level permissions                           │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                   Input Validation                      │    │
│  │  - URL scheme validation (http/https only)              │    │
│  │  - SSRF protection                                      │    │
│  │  - Request body size limits                             │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    Rate Limiting                        │    │
│  │  - Token bucket algorithm                               │    │
│  │  - Per-IP rate limits                                   │    │
│  │  - TTL-based cleanup                                    │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    Audit Logging                        │    │
│  │  - All sensitive operations logged                      │    │
│  │  - Sensitive field filtering                            │    │
│  │  - Tamper-evident storage                               │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

---

### Observability

```
┌─────────────────────────────────────────────────────────────────┐
│                        Logging                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Zap Logger (Structured JSON)                           │    │
│  │  - Request ID correlation                               │    │
│  │  - Log levels (debug, info, warn, error)                │    │
│  │  - Log rotation (lumberjack)                            │    │
│  │  - Sensitive field filtering                            │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                        Metrics                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Prometheus Metrics                                     │    │
│  │  - api_http_requests_total                              │    │
│  │  - api_http_request_duration_seconds                    │    │
│  │  - api_http_requests_in_flight                          │    │
│  │  - test_requests_total                                  │    │
│  │  - test_latency_seconds                                 │    │
│  │  - test_active_workers                                  │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                  │
│                              ▼                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Grafana Dashboards                                     │    │
│  │  - API performance                                      │    │
│  │  - Test execution metrics                               │    │
│  │  - System resources                                     │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                        Tracing                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  OpenTelemetry                                          │    │
│  │  - Distributed trace context propagation                │    │
│  │  - Span creation for handlers                           │    │
│  │  - OTLP export to Jaeger                                │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                  │
│                              ▼                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Jaeger UI                                              │    │
│  │  - Trace visualization                                  │    │
│  │  - Service dependency graph                             │    │
│  │  - Latency analysis                                     │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

---

## Deployment Architecture

### Docker Compose (Development/Staging)

```
┌─────────────────────────────────────────────────────────────────┐
│                     Docker Network                              │
│                                                                 │
│   ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐    │
│   │    Web    │  │    API    │  │  Postgres │  │Prometheus │    │
│   │  :80      │  │  :8080    │  │  :5432    │  │  :9090    │    │
│   │  (nginx)  │  │   (go)    │  │           │  │           │    │
│   └───────────┘  └───────────┘  └───────────┘  └───────────┘    │
│                                                                 │
│   ┌───────────┐  ┌───────────┐                                  │
│   │  Grafana  │  │  Jaeger   │                                  │
│   │  :3000    │  │  :16686   │                                  │
│   │           │  │           │                                  │
│   └───────────┘  └───────────┘                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Production (Kubernetes)

```
┌─────────────────────────────────────────────────────────────────┐
│                     Kubernetes Cluster                          │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    Ingress Controller                   │    │
│  │                (nginx/traefik with TLS)                 │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                  │
│       ┌──────────────────────┼──────────────────────┐           │
│       ▼                      ▼                      ▼           │
│  ┌─────────┐           ┌─────────┐           ┌─────────┐        │
│  │   Web   │           │   API   │           │   API   │        │
│  │ Replica │           │Replica 1│           │Replica 2│        │
│  └─────────┘           └─────────┘           └─────────┘        │
│                              │                                  │
│                              ▼                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │              PostgreSQL (StatefulSet/RDS)               │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

---

## Directory Structure

```
volcanion-stress-test-tool/
├── cmd/
│   ├── server/                 # API server entry point
│   │   └── main.go
│   └── volcanion/              # CLI tool entry point
│       └── main.go
├── internal/                   # Private application code
│   ├── api/
│   │   ├── handler/            # HTTP handlers
│   │   └── router/             # Route configuration
│   ├── audit/                  # Audit logging
│   ├── auth/                   # Authentication (JWT, API keys)
│   ├── config/                 # Configuration management
│   ├── domain/
│   │   ├── models/             # Domain models
│   │   └── service/            # Business logic
│   ├── engine/                 # Load test engine
│   │   ├── scheduler.go        # Test orchestration
│   │   ├── worker.go           # Request execution
│   │   └── metrics.go          # Metrics collection
│   ├── logger/                 # Structured logging
│   ├── metrics/                # Prometheus metrics
│   ├── middleware/             # HTTP middleware
│   ├── reporting/              # Report generation
│   ├── storage/
│   │   └── postgres/           # Database repositories
│   ├── tracing/                # OpenTelemetry setup
│   └── validation/             # Input validation
├── web/                        # React frontend
│   ├── src/
│   │   ├── components/         # UI components
│   │   ├── contexts/           # React contexts
│   │   ├── hooks/              # Custom hooks
│   │   ├── pages/              # Page components
│   │   ├── services/           # API client services
│   │   └── types/              # TypeScript types
│   ├── Dockerfile
│   └── nginx.conf
├── docs/                       # Documentation
├── migrations/                 # Database migrations
├── docker/                     # Docker support files
│   ├── prometheus/
│   └── grafana/
├── .github/workflows/          # CI/CD pipelines
├── Dockerfile                  # Backend Dockerfile
├── docker-compose.yml          # Full stack composition
├── Makefile                    # Build automation
└── go.mod                      # Go module definition
```

---

## Performance Characteristics

| Metric | Value | Notes |
|--------|-------|-------|
| Max concurrent workers | 1000+ | Configurable via `MAX_WORKERS` |
| HTTP connections | 500 idle, 200/host | Shared transport pool |
| Request timeout | 30s default | Per-request configurable |
| WebSocket connections | 1000+ | Per-server |
| Memory usage | ~50MB base | +~1KB per active worker |
| API latency (p99) | <10ms | Without tracing |

---

## Future Considerations

- **Distributed workers**: Support for running workers across multiple nodes
- **Plugin system**: Custom load patterns and assertions
- **GraphQL API**: Alternative to REST for flexible queries
- **Real-time collaboration**: Multiple users monitoring same test
