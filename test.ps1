# Quick Test Script for Volcanion Stress Test Tool

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Testing Volcanion Stress Test Tool" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# Test 1: Health Check
Write-Host "[1] Testing Health Check..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8080/health"
    Write-Host "✓ Health check passed: $($health.status)" -ForegroundColor Green
} catch {
    Write-Host "✗ Health check failed: $_" -ForegroundColor Red
    exit 1
}

Start-Sleep -Seconds 1

# Test 2: Create Test Plan
Write-Host "`n[2] Creating Test Plan..." -ForegroundColor Yellow
$testPlan = @{
    name = "Quick Test"
    target_url = "https://httpbin.org/get"
    method = "GET"
    users = 10
    ramp_up_sec = 2
    duration_sec = 10
    timeout_ms = 5000
} | ConvertTo-Json

try {
    $plan = Invoke-RestMethod -Uri "http://localhost:8080/api/test-plans" -Method POST -Body $testPlan -ContentType "application/json"
    Write-Host "✓ Test plan created successfully" -ForegroundColor Green
    Write-Host "  Plan ID: $($plan.id)" -ForegroundColor Gray
    Write-Host "  Name: $($plan.name)" -ForegroundColor Gray
    Write-Host "  Users: $($plan.users)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Failed to create test plan: $_" -ForegroundColor Red
    exit 1
}

Start-Sleep -Seconds 1

# Test 3: Get All Test Plans
Write-Host "`n[3] Getting All Test Plans..." -ForegroundColor Yellow
try {
    $plans = Invoke-RestMethod -Uri "http://localhost:8080/api/test-plans"
    Write-Host "✓ Retrieved $($plans.Length) test plan(s)" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed to get test plans: $_" -ForegroundColor Red
}

Start-Sleep -Seconds 1

# Test 4: Start Test Run
Write-Host "`n[4] Starting Test Run..." -ForegroundColor Yellow
$startTest = @{
    plan_id = $plan.id
} | ConvertTo-Json

try {
    $run = Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs/start" -Method POST -Body $startTest -ContentType "application/json"
    Write-Host "✓ Test run started successfully" -ForegroundColor Green
    Write-Host "  Run ID: $($run.id)" -ForegroundColor Gray
    Write-Host "  Status: $($run.status)" -ForegroundColor Gray
    Write-Host "  Started: $($run.start_at)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Failed to start test run: $_" -ForegroundColor Red
    exit 1
}

Start-Sleep -Seconds 2

# Test 5: Get Live Metrics
Write-Host "`n[5] Getting Live Metrics..." -ForegroundColor Yellow
try {
    $metrics = Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs/$($run.id)/live"
    Write-Host "✓ Live metrics retrieved" -ForegroundColor Green
    Write-Host "  Total Requests: $($metrics.total_requests)" -ForegroundColor Gray
    Write-Host "  Success: $($metrics.success_requests)" -ForegroundColor Gray
    Write-Host "  Failed: $($metrics.failed_requests)" -ForegroundColor Gray
    Write-Host "  Active Workers: $($metrics.active_workers)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Failed to get live metrics: $_" -ForegroundColor Red
}

# Test 6: Monitor Progress
Write-Host "`n[6] Monitoring Test Progress..." -ForegroundColor Yellow
Write-Host "Checking every 3 seconds for 9 seconds..." -ForegroundColor Gray

for ($i = 1; $i -le 3; $i++) {
    Start-Sleep -Seconds 3
    try {
        $metrics = Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs/$($run.id)/live"
        Write-Host "  [$i] Requests: $($metrics.total_requests) | Success: $($metrics.success_requests) | Failed: $($metrics.failed_requests) | Workers: $($metrics.active_workers)" -ForegroundColor Cyan
    } catch {
        Write-Host "  [$i] Failed to get metrics" -ForegroundColor Red
    }
}

# Wait for test to complete
Write-Host "`n[7] Waiting for test to complete..." -ForegroundColor Yellow
Start-Sleep -Seconds 3

# Test 7: Get Final Metrics
Write-Host "`n[8] Getting Final Metrics..." -ForegroundColor Yellow
try {
    $finalMetrics = Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs/$($run.id)/metrics"
    Write-Host "✓ Final metrics retrieved" -ForegroundColor Green
    Write-Host "`n=== Test Results ===" -ForegroundColor Cyan
    Write-Host "Total Requests: $($finalMetrics.total_requests)" -ForegroundColor White
    Write-Host "Success Rate: $([math]::Round($finalMetrics.success_requests / $finalMetrics.total_requests * 100, 2))%" -ForegroundColor White
    Write-Host "Min Latency: $($finalMetrics.min_latency_ms) ms" -ForegroundColor White
    Write-Host "Avg Latency: $($finalMetrics.avg_latency_ms) ms" -ForegroundColor White
    Write-Host "Max Latency: $($finalMetrics.max_latency_ms) ms" -ForegroundColor White
    Write-Host "P50 Latency: $($finalMetrics.p50_latency_ms) ms" -ForegroundColor White
    Write-Host "P95 Latency: $($finalMetrics.p95_latency_ms) ms" -ForegroundColor White
    Write-Host "P99 Latency: $($finalMetrics.p99_latency_ms) ms" -ForegroundColor White
} catch {
    Write-Host "✗ Failed to get final metrics: $_" -ForegroundColor Red
}

# Test 8: Get Test Run Details
Write-Host "`n[9] Getting Test Run Details..." -ForegroundColor Yellow
try {
    $runDetails = Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs/$($run.id)"
    Write-Host "✓ Test run details retrieved" -ForegroundColor Green
    Write-Host "  Status: $($runDetails.status)" -ForegroundColor Gray
    Write-Host "  Started: $($runDetails.start_at)" -ForegroundColor Gray
    if ($runDetails.end_at) {
        Write-Host "  Ended: $($runDetails.end_at)" -ForegroundColor Gray
    }
} catch {
    Write-Host "✗ Failed to get test run details: $_" -ForegroundColor Red
}

# Test 9: Get All Test Runs
Write-Host "`n[10] Getting All Test Runs..." -ForegroundColor Yellow
try {
    $runs = Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs"
    Write-Host "✓ Retrieved $($runs.Length) test run(s)" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed to get test runs: $_" -ForegroundColor Red
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "All tests completed!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
