# Volcanion Stress Test Tool

A high-performance, web-based API stress testing tool built with Go. This tool allows you to configure, execute, and monitor load tests against HTTP APIs with real-time metrics and Prometheus integration.

## Features

- 🚀 **High Performance**: Built on Go's goroutines for efficient concurrent load generation
- 📊 **Real-time Metrics**: Live monitoring of test execution with percentile calculations (P50, P75, P95, P99)
- 🎯 **Flexible Configuration**: Define test plans with custom headers, body, users, ramp-up time, and duration
- 📈 **Prometheus Integration**: Export metrics to Prometheus for visualization and monitoring
- 🏗️ **Clean Architecture**: Hexagonal architecture for maintainability and extensibility
- 💾 **In-Memory Storage**: Fast in-memory repositories (easily replaceable with database)
- 🔄 **Worker Pool**: Efficient worker pool management with rate control
- 📝 **Structured Logging**: JSON-formatted logs with zap

## Architecture

The project follows Clean Architecture principles with clear separation of concerns:

```
cmd/
  server/
    main.go                 # Application entry point

internal/
  api/
    handler/                # HTTP request handlers
    router/                 # Route configuration
  domain/
    model/                  # Domain models and DTOs
    service/                # Business logic
  engine/
    load_generator.go       # Manages test executions
    scheduler.go            # Coordinates workers
    worker.go               # HTTP request executor
  metrics/
    collector.go            # Prometheus metrics
  storage/
    repository/             # Storage interfaces and implementations
  config/                   # Configuration management
  logger/                   # Logging setup
```

## Prerequisites

- Go >= 1.22
- No external dependencies required (uses in-memory storage)

## Installation

1. Clone the repository:
```powershell
git clone https://github.com/volcanion-company/volcanion-stress-test-tool.git
cd volcanion-stress-test-tool
```

2. Install dependencies:
```powershell
go mod tidy
```

3. Run the server:
```powershell
go run cmd/server/main.go
```

The server will start on port 8080 by default.

## Configuration

Configure the application using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | HTTP server port |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `MAX_WORKERS` | `1000` | Maximum concurrent workers |
| `DEFAULT_TIMEOUT_MS` | `30000` | Default HTTP timeout in milliseconds |
| `METRICS_ENABLED` | `true` | Enable Prometheus metrics |

Example:
```powershell
$env:LOG_LEVEL="debug"
$env:SERVER_PORT="9090"
go run cmd/server/main.go
```

## API Documentation

### Health Check

**GET** `/health`

Returns server health status.

**Response:**
```json
{
  "status": "ok",
  "service": "volcanion-stress-test-tool"
}
```

### Test Plan Endpoints

#### Create Test Plan

**POST** `/api/test-plans`

Create a new test plan configuration.

**Request Body:**
```json
{
  "name": "API Load Test",
  "target_url": "https://api.example.com/endpoint",
  "method": "POST",
  "headers": {
    "Content-Type": "application/json",
    "Authorization": "Bearer token123"
  },
  "body": "{\"key\":\"value\"}",
  "users": 100,
  "ramp_up_sec": 10,
  "duration_sec": 60,
  "timeout_ms": 5000
}
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "API Load Test",
  "target_url": "https://api.example.com/endpoint",
  "method": "POST",
  "headers": {
    "Content-Type": "application/json",
    "Authorization": "Bearer token123"
  },
  "body": "{\"key\":\"value\"}",
  "users": 100,
  "ramp_up_sec": 10,
  "duration_sec": 60,
  "timeout_ms": 5000,
  "created_at": "2025-12-14T10:00:00Z"
}
```

#### Get All Test Plans

**GET** `/api/test-plans`

Retrieve all test plans.

#### Get Test Plan by ID

**GET** `/api/test-plans/:id`

Retrieve a specific test plan.

### Test Run Endpoints

#### Start Test

**POST** `/api/test-runs/start`

Start a new test run based on a test plan.

**Request Body:**
```json
{
  "plan_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response:**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "plan_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "running",
  "start_at": "2025-12-14T10:05:00Z",
  "created_at": "2025-12-14T10:05:00Z"
}
```

#### Stop Test

**POST** `/api/test-runs/:id/stop`

Stop a running test.

**Response:**
```json
{
  "message": "test stopped successfully"
}
```

#### Get All Test Runs

**GET** `/api/test-runs`

Retrieve all test runs.

#### Get Test Run by ID

**GET** `/api/test-runs/:id`

Retrieve a specific test run.

#### Get Test Metrics

**GET** `/api/test-runs/:id/metrics`

Retrieve metrics for a test run (final or stored metrics).

**Response:**
```json
{
  "run_id": "660e8400-e29b-41d4-a716-446655440001",
  "total_requests": 5000,
  "success_requests": 4950,
  "failed_requests": 50,
  "total_duration_ms": 60000,
  "min_latency_ms": 45.2,
  "max_latency_ms": 523.8,
  "avg_latency_ms": 125.3,
  "p50_latency_ms": 112.5,
  "p75_latency_ms": 145.8,
  "p95_latency_ms": 234.2,
  "p99_latency_ms": 387.6,
  "requests_per_sec": 83.33,
  "current_rps": 85.2,
  "active_workers": 100,
  "status_codes": {
    "200": 4950,
    "500": 50
  },
  "errors": {},
  "last_updated": "2025-12-14T10:06:00Z"
}
```

#### Get Live Metrics

**GET** `/api/test-runs/:id/live`

Retrieve real-time metrics for a running test.

### Prometheus Metrics

**GET** `/metrics`

Prometheus-formatted metrics endpoint.

**Metrics Available:**
- `http_request_duration_seconds` - HTTP request latency histogram
- `http_requests_total` - Total number of HTTP requests
- `http_requests_failed_total` - Total number of failed requests
- `stress_test_active_tests` - Number of currently active tests
- `stress_test_active_workers` - Number of currently active workers

## Quick Start Examples

### Example 1: Simple GET Request Test

```powershell
# Create test plan
$testPlan = @{
    name = "Simple GET Test"
    target_url = "https://httpbin.org/get"
    method = "GET"
    users = 50
    ramp_up_sec = 5
    duration_sec = 30
    timeout_ms = 5000
} | ConvertTo-Json

$plan = Invoke-RestMethod -Uri "http://localhost:8080/api/test-plans" -Method POST -Body $testPlan -ContentType "application/json"

# Start test
$startTest = @{
    plan_id = $plan.id
} | ConvertTo-Json

$run = Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs/start" -Method POST -Body $startTest -ContentType "application/json"

# Monitor live metrics
Start-Sleep -Seconds 2
Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs/$($run.id)/live"

# Get final metrics after test completes
Start-Sleep -Seconds 35
Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs/$($run.id)/metrics"
```

### Example 2: POST Request with Headers

```powershell
# Create test plan for POST request
$testPlan = @{
    name = "POST API Test"
    target_url = "https://httpbin.org/post"
    method = "POST"
    headers = @{
        "Content-Type" = "application/json"
        "X-Custom-Header" = "test-value"
    }
    body = '{"message":"Hello, World!","timestamp":1234567890}'
    users = 100
    ramp_up_sec = 10
    duration_sec = 60
    timeout_ms = 5000
} | ConvertTo-Json -Depth 3

$plan = Invoke-RestMethod -Uri "http://localhost:8080/api/test-plans" -Method POST -Body $testPlan -ContentType "application/json"

# Start test
$startTest = @{
    plan_id = $plan.id
} | ConvertTo-Json

$run = Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs/start" -Method POST -Body $startTest -ContentType "application/json"

Write-Host "Test started with ID: $($run.id)"
Write-Host "Monitor at: http://localhost:8080/api/test-runs/$($run.id)/live"
```

### Example 3: Stop a Running Test

```powershell
# Stop test
$runId = "your-run-id-here"
Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs/$runId/stop" -Method POST
```

### Example 4: Using curl (Linux/Mac/Git Bash)

```bash
# Create test plan
curl -X POST http://localhost:8080/api/test-plans \
  -H "Content-Type: application/json" \
  -d '{
    "name": "API Stress Test",
    "target_url": "https://httpbin.org/delay/1",
    "method": "GET",
    "users": 200,
    "ramp_up_sec": 20,
    "duration_sec": 120,
    "timeout_ms": 10000
  }'

# Start test (replace PLAN_ID)
curl -X POST http://localhost:8080/api/test-runs/start \
  -H "Content-Type: application/json" \
  -d '{"plan_id": "PLAN_ID"}'

# Get live metrics (replace RUN_ID)
curl http://localhost:8080/api/test-runs/RUN_ID/live

# Stop test
curl -X POST http://localhost:8080/api/test-runs/RUN_ID/stop
```

## Load Generator Engine Details

### How It Works

1. **Test Plan Creation**: Define the target URL, HTTP method, headers, body, and load parameters
2. **Worker Spawning**: The scheduler spawns workers according to the ramp-up configuration
3. **Request Execution**: Each worker executes HTTP requests continuously until the test duration expires
4. **Metrics Collection**: Workers report latency and status to the centralized metrics collector
5. **Percentile Calculation**: After test completion, percentiles are calculated from all recorded latencies

### Concurrency Model

- **Goroutines**: Each worker runs in its own goroutine
- **Channels**: Used for work distribution and coordination
- **Context**: Manages cancellation and timeout
- **No Blocking**: Non-blocking architecture for maximum throughput

### Rate Control

The tool uses a ticker-based approach for request generation:
- Request channel receives work items at a controlled rate
- Workers pull from the channel when available
- If all workers are busy, requests queue in the channel (buffered)
- This prevents overwhelming the system or the target

## Performance Characteristics

- **Capable of 10,000+ RPS** on modern hardware
- **Low memory footprint** with efficient goroutine management
- **Configurable worker pool** to match system resources
- **HTTP connection pooling** with Keep-Alive for efficiency
- **Percentile calculations** without storing all individual requests in memory (sampled)

## Extending to Distributed Load Testing

To extend this tool for distributed load testing:

### 1. Architecture Changes

```
                    ┌─────────────────┐
                    │  Master Node    │
                    │  (Coordinator)  │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
         ┌────▼────┐    ┌────▼────┐    ┌───▼─────┐
         │ Worker  │    │ Worker  │    │ Worker  │
         │ Node 1  │    │ Node 2  │    │ Node N  │
         └─────────┘    └─────────┘    └─────────┘
```

### 2. Required Changes

**Master Node:**
- Add worker node registration
- Distribute test plans to workers
- Aggregate metrics from all workers
- Coordinate test start/stop

**Worker Node:**
- Register with master
- Receive test plans via RPC/HTTP
- Execute local load generation
- Stream metrics back to master

**Communication:**
- Use gRPC for efficient RPC communication
- Or REST API for simplicity
- Implement heartbeat for worker health monitoring

**Metrics Aggregation:**
- Stream metrics from workers to master
- Aggregate in real-time
- Calculate combined percentiles

### 3. Implementation Steps

1. **Add gRPC definitions**:
```protobuf
service LoadTestService {
  rpc RegisterWorker(WorkerInfo) returns (WorkerRegistration);
  rpc DistributeTest(TestPlan) returns (TestAck);
  rpc StreamMetrics(stream Metrics) returns (MetricsAck);
  rpc StopTest(TestID) returns (StopAck);
}
```

2. **Modify LoadGenerator** to support distributed mode
3. **Add worker discovery** (etcd, Consul, or simple HTTP registry)
4. **Implement metrics aggregation** layer
5. **Add master node API** for cluster management

### 4. Deployment

Use Docker and Kubernetes for distributed deployment:

```yaml
# master-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: stress-test-master
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: master
        image: volcanion-stress-test:latest
        args: ["--mode=master"]
        ports:
        - containerPort: 8080
        - containerPort: 9090  # gRPC

---
# worker-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: stress-test-worker
spec:
  replicas: 10  # Scale workers as needed
  template:
    spec:
      containers:
      - name: worker
        image: volcanion-stress-test:latest
        args: ["--mode=worker", "--master=master-service:9090"]
```

## Troubleshooting

### High Memory Usage

- Reduce `users` count in test plan
- Decrease `duration_sec`
- Adjust `MAX_WORKERS` environment variable

### Connection Errors

- Check `timeout_ms` setting
- Verify target URL is accessible
- Check firewall and network settings
- Increase timeout for slow endpoints

### Low RPS

- Increase `users` count
- Reduce `ramp_up_sec`
- Check if target server is the bottleneck
- Verify network bandwidth

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch
3. Write tests for new features
4. Ensure code follows Go best practices
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Support

For issues and questions:
- GitHub Issues: https://github.com/volcanion-company/volcanion-stress-test-tool/issues
- Documentation: This README

## Roadmap

- [ ] WebSocket support for real-time dashboards
- [ ] Database persistence (PostgreSQL, MongoDB)
- [ ] Authentication and authorization
- [ ] Test result comparison
- [ ] Chain API testing (login → token → authenticated request)
- [ ] Advanced rate limiting (per-second precision)
- [ ] JSON/HTML report export
- [ ] Distributed worker mode
- [ ] Web UI dashboard
- [ ] Request recording and replay

---

Built with ❤️ using Go
