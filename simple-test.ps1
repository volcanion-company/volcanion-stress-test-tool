# Simple Local Test

Write-Host "Starting simple test of Volcanion Stress Test Tool..." -ForegroundColor Cyan

# Create a very simple test plan (10 users, 10 seconds)
Write-Host "`n[1] Creating test plan..." -ForegroundColor Yellow

$testPlan = '{"name":"Simple Local Test","target_url":"https://httpbin.org/get","method":"GET","users":10,"ramp_up_sec":2,"duration_sec":10,"timeout_ms":5000}'

$plan = Invoke-RestMethod -Uri "http://localhost:8080/api/test-plans" -Method POST -Body $testPlan -ContentType "application/json"

Write-Host "Plan created: $($plan.id)" -ForegroundColor Green
Write-Host "  Name: $($plan.name)"
Write-Host "  Target: $($plan.target_url)"
Write-Host "  Users: $($plan.users)"
Write-Host "  Duration: $($plan.duration_sec) seconds"

# Start the test
Write-Host "`n[2] Starting test..." -ForegroundColor Yellow

$startTest = "{`"plan_id`":`"$($plan.id)`"}"
$run = Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs/start" -Method POST -Body $startTest -ContentType "application/json"

Write-Host "Test started: $($run.id)" -ForegroundColor Green
Write-Host "  Status: $($run.status)"
Write-Host "  Started at: $($run.start_at)"

# Wait a moment then check metrics
Write-Host "`n[3] Checking metrics after 3 seconds..." -ForegroundColor Yellow
Start-Sleep -Seconds 3

$metrics = Invoke-RestMethod -Uri "http://localhost:8080/api/test-runs/$($run.id)/live"

Write-Host "Current metrics:" -ForegroundColor Green
Write-Host "  Total requests: $($metrics.total_requests)"
Write-Host "  Success: $($metrics.success_requests)"
Write-Host "  Failed: $($metrics.failed_requests)"
Write-Host "  Active workers: $($metrics.active_workers)"

Write-Host "`nTest is running. Wait 10 seconds for completion..." -ForegroundColor Cyan
Write-Host "Then check final metrics at: http://localhost:8080/api/test-runs/$($run.id)/metrics" -ForegroundColor Gray
