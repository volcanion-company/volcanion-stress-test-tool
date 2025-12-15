# Volcanion Stress Test Tool

<p align="center">
  <img src="docs/logo.png" alt="Volcanion Logo" width="200">
</p>

<p align="center">
  <strong>A powerful, distributed HTTP load testing tool with real-time monitoring</strong>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#installation">Installation</a> •
  <a href="#usage">Usage</a> •
  <a href="#documentation">Documentation</a> •
  <a href="#contributing">Contributing</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/React-18-61DAFB?style=flat&logo=react" alt="React Version">
  <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License">
  <img src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg" alt="PRs Welcome">
</p>

---

## Features

### 🚀 High Performance
- **Concurrent workers** - Scale up to 1000+ concurrent virtual users
- **Connection pooling** - Efficient HTTP client with keep-alive connections
- **Low memory footprint** - Optimized for long-running tests

### 📊 Real-time Monitoring
- **Live metrics dashboard** - Watch test progress in real-time via WebSocket
- **Prometheus integration** - Export metrics for alerting and analysis
- **Grafana dashboards** - Pre-built visualizations for test results

### 🔒 Enterprise Security
- **JWT authentication** - Secure API access with role-based permissions
- **API key support** - Machine-to-machine authentication
- **Rate limiting** - Protect your API from abuse
- **Audit logging** - Track all operations with sensitive data filtering

### 🛠 Developer Experience
- **REST API** - Full-featured API with OpenAPI 3.0 documentation
- **Web UI** - Modern React dashboard for managing tests
- **CLI tool** - Run tests from the command line
- **Docker support** - One-command deployment with docker-compose

### 📈 Observability
- **Structured logging** - JSON logs with request tracing
- **Distributed tracing** - OpenTelemetry integration with Jaeger
- **Health checks** - Kubernetes-ready health endpoints

---

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/volcanion-company/volcanion-stress-test-tool.git
cd volcanion-stress-test-tool

# Start all services
docker-compose up -d

# Access the services
# Web UI:      http://localhost
# API:         http://localhost:8080
# Grafana:     http://localhost:3000 (admin/admin)
# Prometheus:  http://localhost:9090
# Jaeger:      http://localhost:16686
```

### Using Pre-built Binary

```bash
# Download latest release
curl -LO https://github.com/volcanion-company/volcanion-stress-test-tool/releases/latest/download/volcanion-linux-amd64

# Make executable
chmod +x volcanion-linux-amd64

# Run a quick test
./volcanion-linux-amd64 run --url https://httpbin.org/get --users 10 --duration 30s
```

---

## Installation

### Prerequisites

- Go 1.22+ (for building from source)
- Node.js 20+ (for frontend development)
- PostgreSQL 16+ (or use Docker)
- Docker & Docker Compose (optional)

### From Source

```bash
# Clone repository
git clone https://github.com/volcanion-company/volcanion-stress-test-tool.git
cd volcanion-stress-test-tool

# Build backend
make build

# Build frontend
cd web && npm ci && npm run build && cd ..

# Run server
./dist/volcanion-stress-test
```

### Docker Images

```bash
# Build all images
make docker-build-all

# Or pull from registry
docker pull ghcr.io/volcanion-company/volcanion-stress-test:latest
docker pull ghcr.io/volcanion-company/volcanion-stress-test-web:latest
```

---

## Usage

### Web UI

The web interface provides a user-friendly way to:
- Create and manage test plans
- Monitor running tests in real-time
- View historical test results
- Generate reports

Access at `http://localhost` after starting with docker-compose.

### REST API

```bash
# Login and get token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

# Create a test plan
curl -X POST http://localhost:8080/api/v1/test-plans \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "API Load Test",
    "target_url": "https://httpbin.org/get",
    "method": "GET",
    "concurrent_users": 50,
    "duration_seconds": 60,
    "ramp_up_seconds": 10
  }'

# Start a test run
curl -X POST http://localhost:8080/api/v1/test-plans/{plan_id}/run \
  -H "Authorization: Bearer $TOKEN"

# Get live metrics via WebSocket
wscat -c "ws://localhost:8080/api/v1/ws/metrics?token=$TOKEN"
```

### CLI Tool

```bash
# Run a simple test
volcanion run \
  --url https://api.example.com/endpoint \
  --method POST \
  --body '{"key":"value"}' \
  --users 100 \
  --duration 5m \
  --ramp-up 30s

# Run from config file
volcanion run --config test-plan.yaml

# Export results
volcanion report --run-id abc123 --format html --output report.html
```

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ENVIRONMENT` | `development` | Environment (development/staging/production) |
| `SERVER_PORT` | `8080` | API server port |
| `DATABASE_DSN` | - | PostgreSQL connection string |
| `JWT_SECRET` | - | **Required in production** - JWT signing secret |
| `AUTH_ENABLED` | `true` | Enable authentication |
| `RATE_LIMIT_ENABLED` | `true` | Enable rate limiting |
| `RATE_LIMIT_PER_SECOND` | `10` | Requests per second per IP |
| `LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |
| `ALLOWED_ORIGINS` | `localhost` | CORS allowed origins |
| `MAX_WORKERS` | `1000` | Maximum concurrent workers |

### Example Configuration

```yaml
# config.yaml
server:
  port: 8080
  environment: production

database:
  dsn: postgres://user:pass@localhost:5432/volcanion?sslmode=require
  max_conns: 25
  max_idle_conns: 5

auth:
  enabled: true
  jwt_secret: ${JWT_SECRET}
  jwt_duration_hours: 24

rate_limit:
  enabled: true
  requests_per_second: 10

logging:
  level: info
  format: json

tracing:
  enabled: true
  endpoint: http://jaeger:4317
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Web Browser                             │
│                    (React + TanStack Query)                     │
└──────────────────────────┬──────────────────────────────────────┘
                           │ HTTP/WebSocket
┌──────────────────────────┴──────────────────────────────────────┐
│                         Nginx                                   │
│                  (Reverse Proxy + Static)                       │
└──────────────────────────┬──────────────────────────────────────┘
                           │
┌──────────────────────────┴──────────────────────────────────────┐
│                      API Server (Gin)                           │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐    │
│  │  Auth   │ │ CORS    │ │Rate Lim │ │ Logging │ │ Tracing │    │
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘    │
│       └───────────┴───────────┴───────────┴───────────┘         │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    Handlers                              │   │
│  │  TestPlan │ TestRun │ Scenario │ Report │ Auth │ WS      │   │
│  └──────────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    Services                              │   │
│  │  TestService │ ScenarioService │ ReportService           │   │
│  └──────────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                 Load Test Engine                         │   │
│  │  Scheduler │ Workers │ Metrics │ Template                │   │
│  └──────────────────────────────────────────────────────────┘   │
└──────────────────────────┬──────────────────────────────────────┘
                           │
┌──────────────────────────┴──────────────────────────────────────┐
│                      PostgreSQL                                 │
│              (Test Plans, Runs, Results)                        │
└─────────────────────────────────────────────────────────────────┘
```

---

## Documentation

| Document | Description |
|----------|-------------|
| [API Reference](docs/API.md) | REST API documentation |
| [OpenAPI Spec](docs/openapi.yaml) | OpenAPI 3.0 specification |
| [Architecture](docs/ARCHITECTURE.md) | System architecture details |
| [Quick Start](docs/QUICKSTART.md) | Getting started guide |
| [Code Review](docs/review.md) | Code quality assessment |
| [TODO](docs/todo.md) | Roadmap and task tracking |

---

## Development

### Project Structure

```
volcanion-stress-test-tool/
├── cmd/                    # Application entry points
│   ├── server/             # API server
│   └── volcanion/          # CLI tool
├── internal/               # Private application code
│   ├── api/                # REST API handlers & router
│   ├── auth/               # JWT, API keys, passwords
│   ├── config/             # Configuration management
│   ├── domain/             # Domain models & services
│   ├── engine/             # Load test engine
│   ├── middleware/         # HTTP middleware stack
│   ├── storage/            # Database repositories
│   ├── tracing/            # OpenTelemetry setup
│   └── validation/         # Input validation
├── web/                    # React frontend
├── docs/                   # Documentation
├── migrations/             # Database migrations
├── docker/                 # Docker support files
├── .github/workflows/      # CI/CD pipelines
├── Dockerfile              # Backend container
├── docker-compose.yml      # Full stack deployment
└── Makefile                # Build automation
```

### Make Commands

```bash
# Build
make build              # Build Go binary
make build-all          # Build for all platforms
make frontend-build     # Build React frontend

# Test
make test               # Run tests
make test-coverage      # Run with coverage report
make bench              # Run benchmarks

# Lint
make lint               # Run Go linters
make fmt                # Format code

# Docker
make docker-build       # Build backend image
make docker-build-web   # Build frontend image
make docker-build-all   # Build all images
make docker-compose-up  # Start all services
make docker-compose-down # Stop all services

# Development
make dev                # Run in development mode
make frontend-dev       # Run frontend dev server
```

---

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Quick Contribution Steps

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [TanStack Query](https://tanstack.com/query) - Data fetching for React
- [Prometheus](https://prometheus.io/) - Monitoring and alerting
- [OpenTelemetry](https://opentelemetry.io/) - Observability framework
- [Zap](https://github.com/uber-go/zap) - Structured logging

---

<p align="center">
  Made with ❤️ by <a href="https://github.com/volcanion-company">Volcanion Company</a>
</p>
