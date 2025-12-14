# Quick Start Guide

Get started with Volcanion Stress Test Tool in minutes.

## Prerequisites

- Docker & Docker Compose (recommended)
- OR: Go 1.22+, Node.js 20+, PostgreSQL 16+

---

## Option 1: Docker Compose (Recommended)

The fastest way to get started with all services.

### 1. Clone the Repository

```bash
git clone https://github.com/volcanion-company/volcanion-stress-test-tool.git
cd volcanion-stress-test-tool
```

### 2. Start All Services

```bash
docker-compose up -d
```

This starts:
- **Web UI** at http://localhost
- **API Server** at http://localhost:8080
- **PostgreSQL** database
- **Prometheus** at http://localhost:9090
- **Grafana** at http://localhost:3000
- **Jaeger** at http://localhost:16686

### 3. Access the Web UI

Open http://localhost in your browser.

**Default credentials:**
- Username: `admin`
- Password: `admin`

### 4. Create Your First Test

1. Click **"Create New Test"** on the dashboard
2. Fill in the test plan:
   - **Name:** My First Test
   - **Target URL:** https://httpbin.org/get
   - **Method:** GET
   - **Concurrent Users:** 10
   - **Duration:** 60 seconds
3. Click **"Save"**
4. Click **"Run Test"**
5. Watch the real-time metrics!

---

## Option 2: Local Development

For development or when you need more control.

### 1. Clone and Setup

```bash
git clone https://github.com/volcanion-company/volcanion-stress-test-tool.git
cd volcanion-stress-test-tool
```

### 2. Start PostgreSQL

Using Docker:
```bash
docker run -d \
  --name volcanion-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=volcanion \
  -p 5432:5432 \
  postgres:16-alpine
```

### 3. Configure Environment

Create `.env` file:
```bash
cat > .env << EOF
ENVIRONMENT=development
SERVER_PORT=8080
DATABASE_DSN=postgres://postgres:postgres@localhost:5432/volcanion?sslmode=disable
JWT_SECRET=your-secret-key-change-in-production
AUTH_ENABLED=true
LOG_LEVEL=debug
EOF
```

### 4. Run Database Migrations

```bash
# Using psql
psql -h localhost -U postgres -d volcanion -f migrations/001_initial.sql
```

### 5. Start the Backend

```bash
# Build and run
make build
./dist/volcanion-stress-test

# Or run directly
go run cmd/server/main.go
```

### 6. Start the Frontend

```bash
cd web
npm ci
npm run dev
```

Open http://localhost:5173 in your browser.

---

## Option 3: CLI Only

Run tests from the command line without the web UI.

### 1. Download Binary

```bash
# Linux (amd64)
curl -LO https://github.com/volcanion-company/volcanion-stress-test-tool/releases/latest/download/volcanion-linux-amd64
chmod +x volcanion-linux-amd64
sudo mv volcanion-linux-amd64 /usr/local/bin/volcanion

# macOS (arm64)
curl -LO https://github.com/volcanion-company/volcanion-stress-test-tool/releases/latest/download/volcanion-darwin-arm64
chmod +x volcanion-darwin-arm64
sudo mv volcanion-darwin-arm64 /usr/local/bin/volcanion

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/volcanion-company/volcanion-stress-test-tool/releases/latest/download/volcanion-windows-amd64.exe" -OutFile "volcanion.exe"
```

### 2. Run a Simple Test

```bash
volcanion run \
  --url https://httpbin.org/get \
  --method GET \
  --users 10 \
  --duration 30s
```

### 3. Run with Configuration File

Create `test.yaml`:
```yaml
name: API Load Test
target_url: https://api.example.com/users
method: POST
headers:
  Content-Type: application/json
  Authorization: Bearer {{token}}
body: |
  {
    "name": "{{name}}",
    "email": "{{email}}"
  }
concurrent_users: 50
duration_seconds: 300
ramp_up_seconds: 30
variables:
  token: your-api-token
  name: Test User
  email: test@example.com
```

Run:
```bash
volcanion run --config test.yaml
```

---

## First API Test via REST

### 1. Get Authentication Token

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}' | jq -r '.token')

echo "Token: $TOKEN"
```

### 2. Create a Test Plan

```bash
PLAN_ID=$(curl -s -X POST http://localhost:8080/api/v1/test-plans \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Quick API Test",
    "target_url": "https://httpbin.org/post",
    "method": "POST",
    "headers": {
      "Content-Type": "application/json"
    },
    "body": "{\"test\": true}",
    "concurrent_users": 10,
    "duration_seconds": 60
  }' | jq -r '.id')

echo "Plan ID: $PLAN_ID"
```

### 3. Start the Test

```bash
RUN_ID=$(curl -s -X POST "http://localhost:8080/api/v1/test-plans/$PLAN_ID/run" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.run_id')

echo "Run ID: $RUN_ID"
```

### 4. Monitor Progress

```bash
# Get current status
curl -s "http://localhost:8080/api/v1/test-runs/$RUN_ID" \
  -H "Authorization: Bearer $TOKEN" | jq
```

### 5. View Results

```bash
# Wait for completion, then get results
curl -s "http://localhost:8080/api/v1/test-runs/$RUN_ID" \
  -H "Authorization: Bearer $TOKEN" | jq '.metrics'
```

---

## Real-time Metrics via WebSocket

Connect to receive live metrics during test execution:

```javascript
// Browser console or Node.js
const token = 'your-jwt-token';
const runId = 'your-run-id';
const ws = new WebSocket(`ws://localhost:8080/api/v1/ws/metrics?token=${token}&run_id=${runId}`);

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Metrics:', data);
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};
```

---

## Common Use Cases

### Load Testing an API

```bash
volcanion run \
  --url https://api.example.com/endpoint \
  --method GET \
  --users 100 \
  --duration 5m \
  --ramp-up 30s
```

### Stress Testing with POST Data

```bash
volcanion run \
  --url https://api.example.com/users \
  --method POST \
  --header "Content-Type: application/json" \
  --header "Authorization: Bearer token123" \
  --body '{"name": "test", "email": "test@example.com"}' \
  --users 200 \
  --duration 10m
```

### Testing with Think Time

Simulate realistic user behavior with delays between requests:

```bash
volcanion run \
  --url https://api.example.com/endpoint \
  --users 50 \
  --duration 5m \
  --think-time 2s
```

### Testing with Custom Timeouts

```bash
volcanion run \
  --url https://api.example.com/slow-endpoint \
  --users 20 \
  --duration 3m \
  --timeout 60s
```

---

## Viewing Results

### Grafana Dashboard

1. Open http://localhost:3000
2. Login with `admin/admin`
3. Navigate to **Dashboards** → **Volcanion**
4. View real-time and historical metrics

### Jaeger Tracing

1. Open http://localhost:16686
2. Select **volcanion-stress-test** service
3. Click **Find Traces**
4. Analyze request flow and latency

### Prometheus Metrics

1. Open http://localhost:9090
2. Try queries:
   - `test_requests_total`
   - `test_latency_seconds`
   - `api_http_requests_total`

---

## Stopping Services

```bash
# Stop all containers
docker-compose down

# Stop and remove volumes (clean slate)
docker-compose down -v
```

---

## Troubleshooting

### Cannot connect to API

1. Check if services are running:
   ```bash
   docker-compose ps
   ```

2. Check logs:
   ```bash
   docker-compose logs app
   ```

### Authentication errors

1. Ensure `AUTH_ENABLED=true` in environment
2. Check JWT token hasn't expired (24h default)
3. Verify credentials are correct

### Database connection errors

1. Check PostgreSQL is running:
   ```bash
   docker-compose logs postgres
   ```

2. Verify connection string in environment

### High latency results

1. Check target server health
2. Verify network connectivity
3. Consider reducing concurrent users
4. Check if rate limiting is applied

---

## Next Steps

- Read the [API Reference](API_REFERENCE.md) for detailed endpoint documentation
- Explore the [Architecture](ARCHITECTURE.md) to understand the system
- Check out [examples/](../examples/) for more test configurations
- Join our community for support and discussions

---

## Quick Reference

| Component | URL | Credentials |
|-----------|-----|-------------|
| Web UI | http://localhost | admin/admin |
| API | http://localhost:8080 | JWT token |
| Grafana | http://localhost:3000 | admin/admin |
| Prometheus | http://localhost:9090 | - |
| Jaeger | http://localhost:16686 | - |
| PostgreSQL | localhost:5432 | postgres/postgres |
