# Example API Requests for Volcanion Stress Test Tool

This file contains example requests you can use to test the API.

## PowerShell Examples

### 1. Create a Simple GET Test Plan
```powershell
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
Write-Host "Test Plan Created with ID: $($plan.id)"
```

### 2. Create a POST Test Plan with Headers
```powershell
$testPlan = @{
    name = "POST API Test"
    target_url = "https://httpbin.org/post"
    method = "POST"
    headers = @{
        "Content-Type" = "application/json"
        "Authorization" = "Bearer your-token-here"
    }
    body = '{"message":"Hello, World!"}'
    users = 100
    ramp_up_sec = 10
    duration_sec = 60
    timeout_ms = 5000
} | ConvertTo-Json -Depth 3

$plan = Invoke-RestMethod -Uri "http://localhost:8080/api/test-plans" -Method POST -Body $testPlan -ContentType "application/json"
Write-Host "Test Plan Created with ID: $($plan.id)"
```

### 3. Start a Test Run
```powershell
# Replace with your actual plan ID
$planId = "550e8400-e29b-41d4-a716-446655440000"

$startTest = @{
    plan_id = $planId
} | ConvertTo-Json

$run = Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs/start" -Method POST -Body $startTest -ContentType "application/json"
Write-Host "Test Run Started with ID: $($run.id)"
Write-Host "Monitor at: http://localhost:8080/api/test-runs/$($run.id)/live"
```

### 4. Get Live Metrics
```powershell
# Replace with your actual run ID
$runId = "660e8400-e29b-41d4-a716-446655440001"

$metrics = Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs/$runId/live"
Write-Host "Total Requests: $($metrics.total_requests)"
Write-Host "Success: $($metrics.success_requests)"
Write-Host "Failed: $($metrics.failed_requests)"
Write-Host "Avg Latency: $($metrics.avg_latency_ms) ms"
Write-Host "P95 Latency: $($metrics.p95_latency_ms) ms"
Write-Host "Active Workers: $($metrics.active_workers)"
```

### 5. Stop a Running Test
```powershell
# Replace with your actual run ID
$runId = "660e8400-e29b-41d4-a716-446655440001"

Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs/$runId/stop" -Method POST
Write-Host "Test stopped successfully"
```

### 6. Get All Test Plans
```powershell
$plans = Invoke-RestMethod -Uri "http://localhost:8080/api/test-plans"
$plans | ForEach-Object {
    Write-Host "ID: $($_.id) | Name: $($_.name) | Users: $($_.users)"
}
```

### 7. Get All Test Runs
```powershell
$runs = Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs"
$runs | ForEach-Object {
    Write-Host "ID: $($_.id) | Status: $($_.status) | Start: $($_.start_at)"
}
```

### 8. Complete End-to-End Test
```powershell
# Step 1: Create test plan
$testPlan = @{
    name = "E2E Test Example"
    target_url = "https://httpbin.org/delay/1"
    method = "GET"
    users = 20
    ramp_up_sec = 5
    duration_sec = 20
    timeout_ms = 5000
} | ConvertTo-Json

Write-Host "Creating test plan..."
$plan = Invoke-RestMethod -Uri "http://localhost:8080/api/test-plans" -Method POST -Body $testPlan -ContentType "application/json"
Write-Host "Test Plan Created: $($plan.id)"

# Step 2: Start test
Write-Host "`nStarting test run..."
$startTest = @{ plan_id = $plan.id } | ConvertTo-Json
$run = Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs/start" -Method POST -Body $startTest -ContentType "application/json"
Write-Host "Test Run Started: $($run.id)"

# Step 3: Monitor progress
Write-Host "`nMonitoring test (will check every 5 seconds)..."
for ($i = 0; $i -lt 5; $i++) {
    Start-Sleep -Seconds 5
    $metrics = Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs/$($run.id)/live"
    Write-Host "[$i] Requests: $($metrics.total_requests) | Success: $($metrics.success_requests) | Failed: $($metrics.failed_requests) | Workers: $($metrics.active_workers)"
}

# Step 4: Get final metrics
Write-Host "`nWaiting for test to complete..."
Start-Sleep -Seconds 5

$finalMetrics = Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs/$($run.id)/metrics"
Write-Host "`n=== Final Results ==="
Write-Host "Total Requests: $($finalMetrics.total_requests)"
Write-Host "Success Rate: $([math]::Round($finalMetrics.success_requests / $finalMetrics.total_requests * 100, 2))%"
Write-Host "Avg Latency: $($finalMetrics.avg_latency_ms) ms"
Write-Host "P50 Latency: $($finalMetrics.p50_latency_ms) ms"
Write-Host "P95 Latency: $($finalMetrics.p95_latency_ms) ms"
Write-Host "P99 Latency: $($finalMetrics.p99_latency_ms) ms"
Write-Host "Min Latency: $($finalMetrics.min_latency_ms) ms"
Write-Host "Max Latency: $($finalMetrics.max_latency_ms) ms"
Write-Host "RPS: $($finalMetrics.requests_per_sec)"
```

## Bash/Curl Examples

### 1. Create Test Plan
```bash
curl -X POST http://localhost:8080/api/test-plans \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Simple GET Test",
    "target_url": "https://httpbin.org/get",
    "method": "GET",
    "users": 50,
    "ramp_up_sec": 5,
    "duration_sec": 30,
    "timeout_ms": 5000
  }'
```

### 2. Start Test
```bash
# Replace PLAN_ID with actual ID from step 1
curl -X POST http://localhost:8080/api/test-runs/start \
  -H "Content-Type: application/json" \
  -d '{"plan_id": "PLAN_ID"}'
```

### 3. Get Live Metrics
```bash
# Replace RUN_ID with actual ID from step 2
curl http://localhost:8080/api/test-runs/RUN_ID/live | jq
```

### 4. Stop Test
```bash
curl -X POST http://localhost:8080/api/test-runs/RUN_ID/stop
```

### 5. GET All Test Plans
```bash
curl http://localhost:8080/api/test-plans | jq
```

### 6. Get Test Plan by ID
```bash
curl http://localhost:8080/api/test-plans/PLAN_ID | jq
```

### 7. Get All Test Runs
```bash
curl http://localhost:8080/api/test-runs | jq
```

### 8. Get Test Run Metrics
```bash
curl http://localhost:8080/api/test-runs/RUN_ID/metrics | jq
```

### 9. Health Check
```bash
curl http://localhost:8080/health
```

### 10. Prometheus Metrics
```bash
curl http://localhost:8080/metrics
```

## Advanced Examples

### High-Load Test (1000+ Users)
```powershell
$testPlan = @{
    name = "High Load Test"
    target_url = "https://httpbin.org/get"
    method = "GET"
    users = 1000
    ramp_up_sec = 30
    duration_sec = 120
    timeout_ms = 10000
} | ConvertTo-Json

$plan = Invoke-RestMethod -Uri "http://localhost:8080/api/test-plans" -Method POST -Body $testPlan -ContentType "application/json"
$run = Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs/start" -Method POST -Body (@{plan_id=$plan.id} | ConvertTo-Json) -ContentType "application/json"
Write-Host "High load test started: $($run.id)"
```

### API with Authentication
```powershell
$testPlan = @{
    name = "Authenticated API Test"
    target_url = "https://api.example.com/protected"
    method = "GET"
    headers = @{
        "Authorization" = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
        "Accept" = "application/json"
    }
    users = 100
    ramp_up_sec = 10
    duration_sec = 60
    timeout_ms = 5000
} | ConvertTo-Json -Depth 3

$plan = Invoke-RestMethod -Uri "http://localhost:8080/api/test-plans" -Method POST -Body $testPlan -ContentType "application/json"
```

### POST with Complex JSON Body
```powershell
$jsonBody = @{
    user = @{
        name = "John Doe"
        email = "john@example.com"
        age = 30
    }
    preferences = @{
        notifications = $true
        theme = "dark"
    }
    tags = @("developer", "tester")
} | ConvertTo-Json -Depth 5

$testPlan = @{
    name = "Complex POST Test"
    target_url = "https://httpbin.org/post"
    method = "POST"
    headers = @{
        "Content-Type" = "application/json"
    }
    body = $jsonBody
    users = 50
    ramp_up_sec = 5
    duration_sec = 30
    timeout_ms = 5000
} | ConvertTo-Json -Depth 5

$plan = Invoke-RestMethod -Uri "http://localhost:8080/api/test-plans" -Method POST -Body $testPlan -ContentType "application/json"
```

## Notes

- Replace `PLAN_ID` and `RUN_ID` with actual IDs from your responses
- For PowerShell, use `| ConvertTo-Json -Depth 3` for nested objects
- For curl, pipe to `jq` for pretty JSON formatting
- Monitor system resources during high-load tests
- Adjust `users`, `ramp_up_sec`, and `duration_sec` based on your needs and system capacity
