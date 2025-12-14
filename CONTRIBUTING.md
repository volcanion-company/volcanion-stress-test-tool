# Contributing to Volcanion Stress Test Tool

Thank you for your interest in contributing to Volcanion Stress Test Tool! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Code Style](#code-style)
- [Testing](#testing)

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment. Please:

- Be respectful and considerate in all interactions
- Welcome newcomers and help them get started
- Focus on constructive feedback
- Accept responsibility for mistakes and learn from them

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/volcanion-stress-test-tool.git
   cd volcanion-stress-test-tool
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/volcanion-company/volcanion-stress-test-tool.git
   ```

## Development Setup

### Prerequisites

- Go 1.22 or later
- Node.js 20 or later
- PostgreSQL 16 (or use Docker)
- Docker & Docker Compose (optional, for full stack)

### Backend Setup

```bash
# Install Go dependencies
go mod download

# Run linters
make lint

# Run tests
make test

# Build binary
make build

# Run with hot reload (requires air)
air
```

### Frontend Setup

```bash
cd web

# Install dependencies
npm ci

# Run development server
npm run dev

# Build for production
npm run build

# Run linter
npm run lint
```

### Full Stack with Docker

```bash
# Build and start all services
make docker-compose-up-build

# View logs
make docker-compose-logs

# Stop services
make docker-compose-down
```

## Making Changes

1. **Create a branch** from `master`:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. **Make your changes** following our code style guidelines

3. **Write or update tests** as needed

4. **Run the test suite**:
   ```bash
   make test
   make lint
   ```

5. **Commit your changes** following our commit guidelines

## Commit Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/). Format:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `style` | Code style (formatting, etc.) |
| `refactor` | Code refactoring |
| `perf` | Performance improvement |
| `test` | Adding/updating tests |
| `chore` | Build process, dependencies |
| `ci` | CI/CD changes |

### Scopes

| Scope | Description |
|-------|-------------|
| `api` | REST API handlers |
| `engine` | Load test engine |
| `auth` | Authentication |
| `frontend` | React frontend |
| `docker` | Docker/container |
| `db` | Database |
| `config` | Configuration |
| `docs` | Documentation |

### Examples

```bash
feat(engine): add support for custom load patterns
fix(auth): handle expired JWT tokens correctly
docs(api): update OpenAPI specification
test(engine): add benchmarks for scheduler
chore(deps): upgrade Go dependencies
```

## Pull Request Process

1. **Update documentation** if your changes affect it

2. **Ensure CI passes**:
   - All tests pass
   - Linters pass
   - Build succeeds

3. **Create a Pull Request** with:
   - Clear title following commit guidelines
   - Description of changes
   - Link to related issues (if any)
   - Screenshots for UI changes

4. **Address review feedback** promptly

5. **Squash commits** if requested before merging

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows project style
- [ ] Self-review performed
- [ ] Documentation updated
- [ ] No new warnings
```

## Code Style

### Go

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Write meaningful comments for exported functions
- Keep functions small and focused
- Handle errors explicitly

```go
// Good
func ProcessRequest(req *Request) (*Response, error) {
    if req == nil {
        return nil, errors.New("request cannot be nil")
    }
    // ...
}

// Avoid
func process(r *Request) *Response {
    // ...
}
```

### TypeScript/React

- Use TypeScript strict mode
- Follow React best practices
- Use functional components with hooks
- Define proper types (avoid `any`)
- Use meaningful component names

```typescript
// Good
interface UserCardProps {
  user: User;
  onSelect: (userId: string) => void;
}

const UserCard: React.FC<UserCardProps> = ({ user, onSelect }) => {
  // ...
};

// Avoid
const Card = (props: any) => {
  // ...
};
```

## Testing

### Backend Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test -v ./internal/auth/...

# Run benchmarks
make bench
```

### Test Guidelines

- Write table-driven tests
- Mock external dependencies
- Test error cases
- Aim for >80% coverage on critical paths

```go
func TestJWTService_GenerateToken(t *testing.T) {
    tests := []struct {
        name    string
        userID  string
        role    string
        wantErr bool
    }{
        {"valid user", "user-1", "admin", false},
        {"empty userID", "", "admin", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}
```

## Project Structure

```
volcanion-stress-test-tool/
├── cmd/                    # Application entry points
│   ├── server/             # API server
│   └── volcanion/          # CLI tool
├── internal/               # Private application code
│   ├── api/                # REST API (handlers, router)
│   ├── auth/               # Authentication (JWT, API keys)
│   ├── config/             # Configuration
│   ├── domain/             # Domain models & services
│   ├── engine/             # Load test engine
│   ├── middleware/         # HTTP middleware
│   ├── storage/            # Database repositories
│   └── ...
├── web/                    # React frontend
│   └── src/
│       ├── components/     # UI components
│       ├── contexts/       # React contexts
│       ├── hooks/          # Custom hooks
│       ├── pages/          # Page components
│       └── services/       # API services
├── docs/                   # Documentation
├── migrations/             # Database migrations
└── docker/                 # Docker support files
```

## Questions?

Feel free to:
- Open an issue for bugs or feature requests
- Start a discussion for questions
- Reach out to maintainers

Thank you for contributing! 🚀
