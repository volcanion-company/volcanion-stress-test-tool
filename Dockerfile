# ============================================
# Build stage - Go backend
# ============================================
FROM golang:1.22-alpine AS backend-builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build arguments for versioning
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -a -installsuffix cgo \
    -o volcanion-stress-test \
    ./cmd/server

# ============================================
# Build stage - Frontend (optional, if pre-built)
# ============================================
FROM node:20-alpine AS frontend-builder

WORKDIR /app/web

# Copy package files
COPY web/package*.json ./

# Install dependencies
RUN npm ci --only=production=false

# Copy source code
COPY web/ ./

# Build frontend
RUN npm run build

# ============================================
# Final stage - Production image
# ============================================
FROM alpine:3.19

# Labels for container metadata
LABEL org.opencontainers.image.title="Volcanion Stress Test Tool"
LABEL org.opencontainers.image.description="HTTP load testing and stress testing tool"
LABEL org.opencontainers.image.vendor="Volcanion Company"
LABEL org.opencontainers.image.source="https://github.com/volcanion-company/volcanion-stress-test-tool"
LABEL org.opencontainers.image.licenses="MIT"

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    curl \
    && rm -rf /var/cache/apk/*

# Create non-root user for security
RUN addgroup -g 1000 volcanion && \
    adduser -u 1000 -G volcanion -s /bin/sh -D volcanion

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=backend-builder /app/volcanion-stress-test .

# Copy frontend build (if exists)
COPY --from=frontend-builder /app/web/dist ./web/dist

# Copy configuration files
COPY --from=backend-builder /app/docs ./docs

# Create directories for logs and data
RUN mkdir -p /app/logs /app/data && \
    chown -R volcanion:volcanion /app

# Switch to non-root user
USER volcanion

# Expose ports
# 8080 - HTTP API
# 9090 - Prometheus metrics (optional separate port)
EXPOSE 8080

# Environment variables with sensible defaults
ENV SERVER_PORT=8080 \
    LOG_LEVEL=info \
    LOG_FORMAT=json \
    GIN_MODE=release \
    AUTH_ENABLED=true

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the application
ENTRYPOINT ["./volcanion-stress-test"]
