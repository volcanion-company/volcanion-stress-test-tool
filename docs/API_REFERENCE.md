# API Reference

Complete API documentation for Volcanion Stress Test Tool.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

All API endpoints (except `/health` and `/api/v1/auth/login`) require authentication.

### JWT Token

1. Login to get a token:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}'
```

2. Use the token in subsequent requests:
```bash
curl http://localhost:8080/api/v1/test-plans \
  -H "Authorization: Bearer <your-jwt-token>"
```

### API Key

Include your API key in the header:
```bash
curl http://localhost:8080/api/v1/test-plans \
  -H "X-API-Key: <your-api-key>"
```

---

## Endpoints

### Health Check

#### GET /health

Check service health status.

**Response:**
```json
{
  "status": "ok",
  "service": "volcanion-stress-test-tool"
}
```

---

### Authentication

#### POST /api/v1/auth/login

Authenticate and receive a JWT token.

**Request:**
```json
{
  "username": "admin",
  "password": "admin"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-12-15T14:30:00Z",
  "user": {
    "id": "user-123",
    "username": "admin",
    "role": "admin"
  }
}
```

#### POST /api/v1/auth/api-keys

Create a new API key.

**Request:**
```json
{
  "name": "CI/CD Pipeline",
  "expires_in_days": 90
}
```

**Response:**
```json
{
  "id": "key-123",
  "name": "CI/CD Pipeline",
  "key": "vst_abc123def456...",
  "created_at": "2024-12-14T10:00:00Z",
  "expires_at": "2025-03-14T10:00:00Z"
}
```

#### GET /api/v1/auth/api-keys

List all API keys (admin only).

#### DELETE /api/v1/auth/api-keys/{id}

Revoke an API key.

---

### Test Plans

#### GET /api/v1/test-plans

List all test plans.

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | integer | Page number (default: 1) |
| `limit` | integer | Items per page (default: 20) |
| `status` | string | Filter by status |

**Response:**
```json
{
  "data": [
    {
      "id": "plan-123",
      "name": "API Load Test",
      "target_url": "https://api.example.com/users",
      "method": "GET",
      "concurrent_users": 100,
      "duration_seconds": 300,
      "created_at": "2024-12-14T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 1
  }
}
```

#### POST /api/v1/test-plans

Create a new test plan.

**Request:**
```json
{
  "name": "API Load Test",
  "target_url": "https://api.example.com/users",
  "method": "POST",
  "headers": {
    "Content-Type": "application/json",
    "Authorization": "Bearer {{token}}"
  },
  "body": "{\"name\": \"{{name}}\", \"email\": \"{{email}}\"}",
  "concurrent_users": 100,
  "duration_seconds": 300,
  "ramp_up_seconds": 30,
  "think_time_ms": 1000,
  "timeout_ms": 30000,
  "follow_redirects": true,
  "success_status_codes": [200, 201],
  "variables": {
    "token": "abc123",
    "name": "Test User",
    "email": "test@example.com"
  }
}
```

**Response:**
```json
{
  "id": "plan-456",
  "name": "API Load Test",
  "target_url": "https://api.example.com/users",
  "method": "POST",
  "concurrent_users": 100,
  "duration_seconds": 300,
  "created_at": "2024-12-14T10:30:00Z"
}
```

#### GET /api/v1/test-plans/{id}

Get a specific test plan.

#### PUT /api/v1/test-plans/{id}

Update a test plan.

#### DELETE /api/v1/test-plans/{id}

Delete a test plan.

#### POST /api/v1/test-plans/{id}/run

Start a test run from a test plan.

**Response:**
```json
{
  "run_id": "run-789",
  "plan_id": "plan-456",
  "status": "running",
  "started_at": "2024-12-14T10:35:00Z"
}
```

#### POST /api/v1/test-plans/{id}/clone

Clone an existing test plan.

---

### Test Runs

#### GET /api/v1/test-runs

List all test runs.

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `plan_id` | string | Filter by test plan |
| `status` | string | Filter by status (running, completed, failed, cancelled) |
| `from` | datetime | Start date filter |
| `to` | datetime | End date filter |

**Response:**
```json
{
  "data": [
    {
      "id": "run-789",
      "plan_id": "plan-456",
      "status": "completed",
      "start_at": "2024-12-14T10:35:00Z",
      "end_at": "2024-12-14T10:40:00Z"
    }
  ]
}
```

#### GET /api/v1/test-runs/{id}

Get test run details including metrics.

**Response:**
```json
{
  "id": "run-789",
  "plan_id": "plan-456",
  "status": "completed",
  "start_at": "2024-12-14T10:35:00Z",
  "end_at": "2024-12-14T10:40:00Z",
  "metrics": {
    "total_requests": 15000,
    "successful_requests": 14850,
    "failed_requests": 150,
    "avg_latency_ms": 125.5,
    "min_latency_ms": 45,
    "max_latency_ms": 2500,
    "p50_latency_ms": 110,
    "p95_latency_ms": 350,
    "p99_latency_ms": 890,
    "requests_per_second": 50.2,
    "status_codes": {
      "200": 14000,
      "201": 850,
      "500": 150
    }
  }
}
```

#### POST /api/v1/test-runs/{id}/stop

Stop a running test.

**Response:**
```json
{
  "id": "run-789",
  "status": "cancelled",
  "stopped_at": "2024-12-14T10:38:00Z"
}
```

#### GET /api/v1/test-runs/{id}/metrics

Get real-time metrics for a running test.

---

### Scenarios

Multi-step test scenarios for complex workflows.

#### GET /api/v1/scenarios

List all scenarios.

#### POST /api/v1/scenarios

Create a new scenario.

**Request:**
```json
{
  "name": "User Registration Flow",
  "description": "Test user signup and login flow",
  "steps": [
    {
      "name": "Register User",
      "method": "POST",
      "url": "https://api.example.com/register",
      "headers": {
        "Content-Type": "application/json"
      },
      "body": "{\"email\": \"{{email}}\", \"password\": \"{{password}}\"}",
      "extract": {
        "user_id": "$.id",
        "token": "$.token"
      },
      "assertions": [
        {"type": "status", "value": 201},
        {"type": "json_path", "path": "$.id", "exists": true}
      ]
    },
    {
      "name": "Get User Profile",
      "method": "GET",
      "url": "https://api.example.com/users/{{user_id}}",
      "headers": {
        "Authorization": "Bearer {{token}}"
      },
      "assertions": [
        {"type": "status", "value": 200},
        {"type": "response_time", "max_ms": 500}
      ]
    }
  ],
  "variables": {
    "email": "test{{$random}}@example.com",
    "password": "securepassword123"
  }
}
```

#### GET /api/v1/scenarios/{id}

Get scenario details.

#### PUT /api/v1/scenarios/{id}

Update a scenario.

#### DELETE /api/v1/scenarios/{id}

Delete a scenario.

#### POST /api/v1/scenarios/{id}/run

Execute a scenario.

---

### Reports

#### GET /api/v1/reports

List generated reports.

#### POST /api/v1/reports

Generate a new report.

**Request:**
```json
{
  "run_id": "run-789",
  "format": "html",
  "include_charts": true,
  "include_details": true
}
```

**Response:**
```json
{
  "id": "report-123",
  "run_id": "run-789",
  "format": "html",
  "status": "generating",
  "created_at": "2024-12-14T11:00:00Z"
}
```

#### GET /api/v1/reports/{id}

Get report details.

#### GET /api/v1/reports/{id}/download

Download report file.

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `format` | string | Output format (html, pdf, json, csv) |

---

### Audit Logs

#### GET /api/v1/audit-logs

List audit logs (admin only).

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `user_id` | string | Filter by user |
| `action` | string | Filter by action type |
| `from` | datetime | Start date |
| `to` | datetime | End date |

**Response:**
```json
{
  "data": [
    {
      "id": "log-123",
      "user_id": "user-456",
      "action": "test_plan.create",
      "resource_type": "test_plan",
      "resource_id": "plan-789",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "timestamp": "2024-12-14T10:30:00Z"
    }
  ]
}
```

---

## WebSocket API

### Real-time Metrics

Connect to receive live metrics during test execution.

**Endpoint:** `ws://localhost:8080/api/v1/ws/metrics`

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `token` | string | JWT token for authentication |
| `run_id` | string | Test run ID to subscribe to |

**Example:**
```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/ws/metrics?token=<jwt>&run_id=run-789');

ws.onmessage = (event) => {
  const metrics = JSON.parse(event.data);
  console.log('Live metrics:', metrics);
};
```

**Message Format:**
```json
{
  "type": "metrics",
  "run_id": "run-789",
  "timestamp": "2024-12-14T10:36:00Z",
  "data": {
    "total_requests": 5000,
    "successful_requests": 4950,
    "current_rps": 48.5,
    "active_workers": 100,
    "avg_latency_ms": 122.3,
    "p95_latency_ms": 340,
    "status_codes": {"200": 4950, "500": 50}
  }
}
```

### Test Events

**Endpoint:** `ws://localhost:8080/api/v1/ws/events`

Receive test lifecycle events.

**Event Types:**
- `test.started` - Test run started
- `test.completed` - Test run completed
- `test.failed` - Test run failed
- `test.cancelled` - Test run cancelled
- `metrics.update` - Metrics snapshot

---

## Error Handling

All errors follow a consistent format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request body",
    "details": [
      {
        "field": "target_url",
        "message": "must be a valid URL"
      }
    ]
  },
  "request_id": "req-abc123"
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `UNAUTHORIZED` | 401 | Missing or invalid authentication |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 400 | Invalid request data |
| `RATE_LIMITED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Server error |

---

## Rate Limiting

API requests are rate-limited per IP address:

- **Default:** 10 requests/second
- **Burst:** Up to 50 requests

Rate limit headers:
```
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 8
X-RateLimit-Reset: 1702558200
```

---

## Pagination

List endpoints support pagination:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number |
| `limit` | integer | 20 | Items per page (max: 100) |

Response includes pagination info:
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

---

## Request Headers

| Header | Required | Description |
|--------|----------|-------------|
| `Authorization` | Yes* | Bearer token: `Bearer <jwt>` |
| `X-API-Key` | Yes* | Alternative to Authorization |
| `Content-Type` | Yes | `application/json` for POST/PUT |
| `X-Request-ID` | No | Custom request ID for tracing |

*One of `Authorization` or `X-API-Key` is required.

---

## Response Headers

| Header | Description |
|--------|-------------|
| `X-Request-ID` | Unique request identifier |
| `X-API-Version` | API version |
| `X-RateLimit-*` | Rate limit information |

---

## OpenAPI Specification

Full OpenAPI 3.0 specification available at:
- **Swagger UI:** http://localhost:8080/api/docs/
- **JSON:** http://localhost:8080/api/docs/openapi.json
- **YAML:** [docs/openapi.yaml](openapi.yaml)
