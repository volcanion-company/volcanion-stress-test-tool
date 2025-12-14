# Multi-stage build for smaller image size
FROM golang:1.22-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o volcanion-stress-test cmd/server/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/volcanion-stress-test .

# Expose port
EXPOSE 8080

# Set environment variables
ENV SERVER_PORT=8080
ENV LOG_LEVEL=info

# Run the application
CMD ["./volcanion-stress-test"]
