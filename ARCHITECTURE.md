# Volcanion Stress Test Tool - Architecture Documentation

## Overview

This document describes the architecture and design decisions of the Volcanion Stress Test Tool.

## Architecture Pattern

The application follows **Clean Architecture** (also known as Hexagonal Architecture) principles:

```
┌─────────────────────────────────────────────────────────┐
│                     Presentation Layer                   │
│              (HTTP Handlers, Router, API)                │
└───────────────────────┬─────────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────┐
│                     Service Layer                        │
│              (Business Logic, Orchestration)             │
└───────────────────────┬─────────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────┐
│                     Domain Layer                         │
│         (Models, Entities, Core Business Rules)          │
└───────────────────────┬─────────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────┐
│                  Infrastructure Layer                    │
│     (Repositories, External Services, Load Engine)       │
└─────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### 1. Presentation Layer (`/internal/api`)

**Purpose**: Handle HTTP requests and responses

**Components**:
- **Handlers**: Process HTTP requests, validate input, call services
- **Router**: Define API routes and middleware

**Key Files**:
- `handler/test_plan_handler.go`
- `handler/test_run_handler.go`
- `router/router.go`

**Responsibilities**:
- Request validation
- Response formatting
- HTTP-specific error handling
- Route configuration

### 2. Service Layer (`/internal/domain/service`)

**Purpose**: Implement business logic and orchestrate operations

**Components**:
- **TestService**: Manages test lifecycle, coordinates repositories and engine

**Key Files**:
- `service/test_service.go`

**Responsibilities**:
- Business rule enforcement
- Transaction coordination
- Cross-cutting concerns (logging, metrics)
- Background task management

### 3. Domain Layer (`/internal/domain/model`)

**Purpose**: Define core business entities and rules

**Components**:
- **Models**: TestPlan, TestRun, Metrics
- **DTOs**: Request/Response objects

**Key Files**:
- `model/test.go`
- `model/metrics.go`

**Responsibilities**:
- Domain object definitions
- Business invariants
- Domain logic (validation, calculations)

### 4. Infrastructure Layer

**Purpose**: Implement technical capabilities

**Components**:
- **Storage** (`/internal/storage/repository`): Data persistence
- **Engine** (`/internal/engine`): Load generation
- **Metrics** (`/internal/metrics`): Observability
- **Logger** (`/internal/logger`): Logging
- **Config** (`/internal/config`): Configuration

**Key Files**:
- `repository/repository.go` (interfaces)
- `repository/memory_repo.go` (implementation)
- `engine/load_generator.go`
- `engine/scheduler.go`
- `engine/worker.go`

## Load Generation Architecture

### Concurrency Model

```
                    ┌──────────────┐
                    │  Scheduler   │
                    └──────┬───────┘
                           │
                  ┌────────▼────────┐
                  │ Request Channel │ (buffered)
                  └────────┬────────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
   ┌────▼────┐       ┌─────▼────┐      ┌─────▼────┐
   │ Worker  │       │ Worker   │      │ Worker   │
   │   #1    │       │   #2     │ ...  │   #N     │
   └────┬────┘       └─────┬────┘      └─────┬────┘
        │                  │                  │
        └──────────────────┼──────────────────┘
                           │
                    ┌──────▼───────┐
                    │   Metrics    │
                    │  Collector   │
                    └──────────────┘
```

### Components

#### Scheduler
- Manages worker lifecycle
- Implements ramp-up logic
- Controls test duration
- Reports metrics periodically

#### Worker
- Executes HTTP requests
- Records latency
- Handles errors
- Uses HTTP connection pooling

#### Request Channel
- Distributes work to workers
- Implements backpressure
- Non-blocking architecture

#### Metrics Collector
- Thread-safe metrics aggregation
- Real-time statistics
- Percentile calculation

### Flow

1. **Initialization**:
   - Create test plan
   - Initialize scheduler with plan parameters
   - Create request channel

2. **Ramp-Up**:
   - Spawn workers gradually
   - Increase rate over time
   - Update active worker count

3. **Steady State**:
   - All workers active
   - Constant request rate
   - Continuous metrics collection

4. **Completion**:
   - Context cancelled after duration
   - Workers finish current requests
   - Calculate final percentiles
   - Update test run status

## Data Flow

### Creating and Running a Test

```
1. Client → POST /api/test-plans
              ↓
2. Handler validates request
              ↓
3. Service creates TestPlan entity
              ↓
4. Repository stores TestPlan
              ↓
5. Return TestPlan to client

6. Client → POST /api/test-runs/start
              ↓
7. Service retrieves TestPlan
              ↓
8. Create TestRun entity
              ↓
9. Start LoadGenerator
              ↓
10. Scheduler spawns workers
              ↓
11. Workers execute requests
              ↓
12. Metrics collected in real-time
              ↓
13. Background goroutine monitors completion
              ↓
14. Update TestRun status when done
```

### Retrieving Live Metrics

```
1. Client → GET /api/test-runs/:id/live
              ↓
2. Handler calls service
              ↓
3. Service checks if test is running
              ↓
4. If running: Get from LoadGenerator
   If not: Get from Repository
              ↓
5. Return Metrics snapshot
```

## Concurrency & Thread Safety

### Synchronization Mechanisms

1. **Mutexes** (sync.RWMutex):
   - Metrics: Protects concurrent updates
   - Repositories: Protects map access
   - LoadGenerator: Protects active tests map

2. **Channels**:
   - Request distribution
   - Worker coordination
   - Graceful shutdown

3. **Context**:
   - Test cancellation
   - Timeout management
   - Propagation to workers

### Thread-Safe Operations

- All repository operations use RWMutex
- Metrics updates are synchronized
- Snapshot operations create copies
- No shared mutable state between workers

## Metrics & Observability

### Internal Metrics

Stored in `Metrics` model:
- Request counts (total, success, failed)
- Latency statistics (min, max, avg, percentiles)
- Active worker count
- Status code distribution
- Error types

### Prometheus Metrics

Exposed at `/metrics`:
- `http_request_duration_seconds` (Histogram)
- `http_requests_total` (Counter)
- `http_requests_failed_total` (Counter)
- `stress_test_active_tests` (Gauge)
- `stress_test_active_workers` (Gauge)

### Logging

Structured JSON logs with zap:
- Test lifecycle events
- Error conditions
- Performance metrics (every 5 seconds)
- Worker activities (debug level)

## Scalability Considerations

### Vertical Scaling

The current architecture supports:
- **CPU**: Each worker runs in a goroutine (lightweight)
- **Memory**: Efficient data structures, metrics sampling
- **Network**: Connection pooling, Keep-Alive

Tested for 10,000+ concurrent users on modern hardware.

### Horizontal Scaling (Future)

For distributed load testing:

1. **Master-Worker Pattern**:
   - Master coordinates multiple worker nodes
   - Workers execute load locally
   - Metrics aggregated at master

2. **Communication**:
   - gRPC for efficient RPC
   - Streaming for real-time metrics

3. **Discovery**:
   - Service registry (etcd, Consul)
   - Dynamic worker registration

4. **Deployment**:
   - Kubernetes for orchestration
   - StatefulSets for workers
   - Horizontal Pod Autoscaler

## Design Decisions

### 1. In-Memory Storage

**Decision**: Use in-memory maps for storage

**Rationale**:
- Fast access for real-time metrics
- Simple to implement and test
- Easy to replace with database

**Trade-offs**:
- Data lost on restart
- Limited by memory
- No persistence

**Future**: Add database adapter (PostgreSQL, MongoDB)

### 2. Goroutine-Based Workers

**Decision**: One goroutine per worker

**Rationale**:
- Lightweight (2KB stack)
- Efficient context switching
- Native Go concurrency

**Trade-offs**:
- Limited by GOMAXPROCS
- Requires proper synchronization

### 3. Channel-Based Work Distribution

**Decision**: Use buffered channel for request distribution

**Rationale**:
- Non-blocking producer
- Natural backpressure
- Simple coordination

**Trade-offs**:
- Memory for buffer
- Potential latency in queue

### 4. Percentile Calculation

**Decision**: Calculate percentiles post-test from collected samples

**Rationale**:
- Accurate percentiles
- Simple implementation
- Memory efficient (only workers store latencies)

**Trade-offs**:
- Not available in real-time
- Memory increases with duration

**Alternative**: Use approximation algorithms (t-digest, HdrHistogram)

### 5. HTTP Client Configuration

**Decision**: Custom HTTP client per worker with connection pooling

**Rationale**:
- Reuse connections (Keep-Alive)
- Control timeout per worker
- Independent from Go default client

**Configuration**:
```go
client := &http.Client{
    Timeout: time.Duration(plan.TimeoutMs) * time.Millisecond,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 100,
        IdleConnTimeout:     90 * time.Second,
    },
}
```

## Error Handling

### Levels

1. **Request Level**: Worker records error, continues
2. **Worker Level**: Worker fails, test continues
3. **Scheduler Level**: Log error, may stop test
4. **Service Level**: Return error to client

### Strategy

- Fail fast for configuration errors
- Graceful degradation for runtime errors
- All errors logged with context
- Metrics include error counts by type

## Testing Strategy

### Unit Tests

Test individual components:
- Models (validation, serialization)
- Repository operations
- Service business logic

### Integration Tests

Test component interaction:
- API endpoints
- Service → Repository
- Engine → Metrics

### Load Tests

Test performance:
- Maximum concurrent users
- Sustained throughput
- Memory usage under load

## Security Considerations

### Current

- No authentication (internal tool)
- Input validation on all endpoints
- No data persistence (reduced attack surface)

### Future Enhancements

- API key authentication
- Rate limiting per client
- TLS/HTTPS support
- Role-based access control
- Request signing for distributed mode

## Performance Optimization

### Current Optimizations

1. **HTTP Connection Pooling**: Reuse connections
2. **Buffered Channels**: Reduce synchronization overhead
3. **RWMutex**: Allow concurrent reads
4. **Efficient Serialization**: Use Gin's optimized JSON
5. **Worker Pool**: Limit goroutine creation

### Future Optimizations

1. **Zero-Copy Metrics**: Use atomic operations
2. **Batch Metrics Updates**: Reduce lock contention
3. **Custom HTTP Client**: Optimize for stress testing
4. **Memory Pooling**: Reuse request/response buffers
5. **CPU Pinning**: Dedicated cores for workers

## Monitoring & Debugging

### Metrics to Monitor

- `active_workers`: Should match test plan
- `requests_total`: Should increase linearly
- `request_duration`: Watch for anomalies
- `failed_requests`: Should be low

### Debugging Tools

1. **Logs**: Check structured logs for errors
2. **Prometheus**: Visualize metrics in Grafana
3. **pprof**: Go profiling for performance issues
4. **Live Metrics**: Real-time endpoint for monitoring

### Common Issues

1. **High latency**: Check target server, network
2. **Low RPS**: Increase workers, check CPU
3. **Errors**: Check target availability, timeouts
4. **Memory growth**: Check test duration, worker count

## Future Enhancements

See README.md Roadmap section for planned features.

---

This architecture provides a solid foundation for high-performance stress testing with clear separation of concerns and extensibility for future enhancements.
