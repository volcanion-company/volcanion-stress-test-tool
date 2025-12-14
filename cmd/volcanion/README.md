# Volcanion CLI

Command-line interface for Volcanion Stress Test Tool.

## Installation

```bash
cd cmd/volcanion
go build -o volcanion.exe
```

## Quick Start

```bash
# Run test from file with live monitoring
./volcanion run -f ../../examples/simple-load-test.yaml --watch

# List recent test runs
./volcanion list runs --limit 10

# Export results as HTML
./volcanion export <run-id> --format html -o report.html
```

## Commands

### run

Execute a stress test from YAML/JSON file or existing plan ID.

```bash
# Run from file
./volcanion run -f test-plan.yaml

# Run with live monitoring
./volcanion run -f test-plan.yaml --watch

# Save results to file
./volcanion run -f test-plan.yaml -o results.json

# Run from existing plan ID
./volcanion run --plan-id abc123

# Disable colored output
./volcanion run -f test-plan.yaml --no-color
```

**Flags:**
- `-f, --file` - Path to YAML/JSON test plan file
- `--plan-id` - Existing test plan ID to run
- `-w, --watch` - Watch test progress in real-time
- `-o, --output` - Save results to file
- `--no-color` - Disable colored output

**Live Monitoring:**

When using `--watch`, displays:
- Progress bar showing elapsed time
- Current metrics (requests, RPS, latency)
- Success/failure counts
- Real-time updates every 2 seconds

Press Ctrl+C to stop watching (test continues running).

### list

List test plans or test runs.

```bash
# List all test plans
./volcanion list plans

# List all test runs
./volcanion list runs

# Filter by status
./volcanion list runs --status running
./volcanion list runs --status completed
./volcanion list runs --status failed

# Limit results
./volcanion list runs --limit 10
```

**Flags:**
- `--status` - Filter runs by status (running, completed, failed)
- `--limit` - Limit number of results

**Output:**

Table format with columns:
- ID
- Name
- Status
- Start/End time
- Duration

### export

Export test results to various formats.

```bash
# Export as JSON
./volcanion export <run-id> --format json -o results.json

# Export as CSV
./volcanion export <run-id> --format csv -o metrics.csv

# Export as HTML report
./volcanion export <run-id> --format html -o report.html
```

**Flags:**
- `-f, --format` - Export format: json, csv, html (default: json)
- `-o, --output` - Output file path (required)

**Export Formats:**

**JSON**: Complete data structure
```json
{
  "test_run": {...},
  "metrics": {...}
}
```

**CSV**: Spreadsheet-friendly metrics
```csv
Timestamp,Metric,Value
2025-12-14T12:00:00Z,TotalRequests,1000
2025-12-14T12:00:00Z,SuccessfulRequests,990
```

**HTML**: Beautiful report with charts and styling
- Summary statistics
- Detailed metrics
- Printable format
- Embedded CSS

## Global Flags

Available for all commands:

- `--api` - API base URL (default: http://localhost:8080)
- `--config` - Config file path (default: ~/.volcanion.yaml)
- `-v, --verbose` - Enable verbose logging
- `-h, --help` - Show help

## Configuration

### Config File

Create `~/.volcanion.yaml`:

```yaml
api: http://localhost:8080
verbose: true
```

### Environment Variables

```bash
export VOLCANION_API=http://production-server:8080
export VOLCANION_VERBOSE=true
export VOLCANION_CONFIG=/path/to/config.yaml
```

### Precedence

1. Command-line flags (highest priority)
2. Environment variables
3. Config file
4. Default values (lowest priority)

## Test Plan Format

### YAML Example

```yaml
name: "API Load Test"
target_url: "https://api.example.com/users"
method: "GET"

# Load configuration
concurrent_users: 50
duration_sec: 120
ramp_up_sec: 10
timeout_ms: 5000

# HTTP settings
headers:
  Content-Type: "application/json"
  Authorization: "Bearer token123"

# Optional: Request body
body: |
  {
    "key": "value"
  }

# Load pattern
rate_pattern: "fixed"
target_rps: 100

# Optional: SLA thresholds
sla_config:
  max_avg_latency_ms: 200
  max_p95_latency_ms: 500
  max_p99_latency_ms: 1000
  min_success_rate: 99.0
```

### JSON Example

```json
{
  "name": "API Load Test",
  "target_url": "https://api.example.com/users",
  "method": "POST",
  "concurrent_users": 50,
  "duration_sec": 120,
  "ramp_up_sec": 10,
  "timeout_ms": 5000,
  "headers": {
    "Content-Type": "application/json"
  },
  "body": "{\"key\":\"value\"}",
  "rate_pattern": "fixed",
  "target_rps": 100
}
```

## Load Patterns

### Fixed

Constant request rate:

```yaml
rate_pattern: "fixed"
target_rps: 100
```

### Ramp

Gradually increase load:

```yaml
rate_pattern: "ramp"
rate_steps:
  - duration_sec: 60
    target_rps: 10
  - duration_sec: 60
    target_rps: 100
```

### Step

Step-wise changes:

```yaml
rate_pattern: "step"
rate_steps:
  - duration_sec: 30
    target_rps: 50
  - duration_sec: 30
    target_rps: 100
  - duration_sec: 30
    target_rps: 200
```

### Spike

Sudden bursts:

```yaml
rate_pattern: "spike"
rate_steps:
  - duration_sec: 60
    target_rps: 50
  - duration_sec: 10
    target_rps: 500
  - duration_sec: 60
    target_rps: 50
```

## Examples

### Basic GET Request

```yaml
name: "Simple GET Test"
target_url: "https://httpbin.org/get"
method: "GET"
concurrent_users: 10
duration_sec: 60
rate_pattern: "fixed"
target_rps: 50
```

### POST with Authentication

```yaml
name: "Authenticated POST"
target_url: "https://api.example.com/data"
method: "POST"
concurrent_users: 20
duration_sec: 120
headers:
  Content-Type: "application/json"
  Authorization: "Bearer your-token"
body: |
  {
    "name": "test",
    "value": 123
  }
rate_pattern: "ramp"
rate_steps:
  - duration_sec: 30
    target_rps: 10
  - duration_sec: 90
    target_rps: 50
```

### Spike Test with SLA

```yaml
name: "Spike Test"
target_url: "https://api.example.com/endpoint"
method: "GET"
concurrent_users: 100
duration_sec: 180
rate_pattern: "spike"
rate_steps:
  - duration_sec: 60
    target_rps: 50
  - duration_sec: 20
    target_rps: 500
  - duration_sec: 100
    target_rps: 50
sla_config:
  max_avg_latency_ms: 300
  max_p95_latency_ms: 800
  min_success_rate: 95.0
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Performance Test

on: [push]

jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Setup Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.22
      
      - name: Build CLI
        run: |
          cd cmd/volcanion
          go build -o volcanion
      
      - name: Run Load Test
        run: |
          ./cmd/volcanion/volcanion run \
            -f tests/load-test.yaml \
            -o results.json
      
      - name: Upload Results
        uses: actions/upload-artifact@v2
        with:
          name: test-results
          path: results.json
```

### GitLab CI

```yaml
performance:
  stage: test
  script:
    - cd cmd/volcanion
    - go build -o volcanion
    - ./volcanion run -f ../../tests/load-test.yaml -o results.json
  artifacts:
    paths:
      - results.json
    reports:
      performance: results.json
```

### Jenkins

```groovy
pipeline {
    agent any
    stages {
        stage('Build') {
            steps {
                sh 'cd cmd/volcanion && go build -o volcanion'
            }
        }
        stage('Load Test') {
            steps {
                sh './cmd/volcanion/volcanion run -f tests/load-test.yaml -o results.json'
            }
        }
        stage('Archive Results') {
            steps {
                archiveArtifacts artifacts: 'results.json'
            }
        }
    }
}
```

## Output Examples

### Run Command (Normal)

```
✓ Test plan created: abc123
✓ Test started: run-456
⏳ Waiting for test to complete...
✓ Test completed successfully

Test Results:
  Total Requests:    12,500
  Successful:        12,450 (99.6%)
  Failed:            50 (0.4%)
  Avg Latency:       145.23 ms
  P95 Latency:       234.56 ms
  P99 Latency:       456.78 ms
  Throughput:        104.17 req/s
```

### Run Command (Watch Mode)

```
ℹ Watching live metrics... (Press Ctrl+C to stop watching)

Progress [=========>              ] 45s / 120s

  Requests:     4,500 | Success: 4,475 (99.4%) | Failed:     25
  RPS:          100.0 req/s
  Avg Latency:  145.23 ms
  P95 Latency:  234.56 ms
  P99 Latency:  456.78 ms

✓ Test completed
```

### List Command

```
Test Plans:
ID                  Name              Created At
abc123              API Load Test     2025-12-14 10:30:00
def456              Spike Test        2025-12-14 11:15:00
ghi789              Endurance Test    2025-12-14 12:00:00

Total: 3 plans
```

### Export Command

```
✓ Exported test results to report.html
  Format: HTML
  Size: 45.6 KB
```

## Troubleshooting

### Connection Refused

```bash
# Ensure backend is running
# Check API URL
./volcanion run -f test.yaml --api http://localhost:8080
```

### Invalid Test Plan

```bash
# Validate YAML syntax
cat test-plan.yaml

# Check required fields
# - name
# - target_url
# - method
# - concurrent_users
# - duration_sec
```

### Permission Denied

```bash
# Make executable
chmod +x volcanion

# Or run with go
go run main.go run -f test.yaml
```

### Config File Not Found

```bash
# Create config file
cat > ~/.volcanion.yaml <<EOF
api: http://localhost:8080
verbose: true
EOF

# Or specify path
./volcanion run -f test.yaml --config /path/to/config.yaml
```

## Advanced Usage

### Script Integration

```bash
#!/bin/bash

# Run test and check exit code
if ./volcanion run -f test.yaml -o results.json; then
    echo "Test passed"
    ./volcanion export <run-id> --format html -o report.html
else
    echo "Test failed"
    exit 1
fi
```

### Parallel Tests

```bash
# Run multiple tests in parallel
./volcanion run -f test1.yaml &
./volcanion run -f test2.yaml &
./volcanion run -f test3.yaml &
wait
```

### Result Processing

```bash
# Export and process results
./volcanion export <run-id> -f json -o results.json
cat results.json | jq '.metrics.avg_latency_ms'
```

## Dependencies

- **Cobra** v1.10.2 - CLI framework
- **Viper** v1.21.0 - Configuration
- **Color** v1.18.0 - Terminal colors
- **Progressbar** v3.18.0 - Progress indicators
- **YAML** v3 - YAML parsing

## Building

### Standard Build

```bash
go build -o volcanion
```

### Optimized Build

```bash
go build -ldflags="-s -w" -o volcanion
```

### Cross-Platform

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o volcanion.exe

# Linux
GOOS=linux GOARCH=amd64 go build -o volcanion

# macOS
GOOS=darwin GOARCH=amd64 go build -o volcanion
```

## License

MIT
