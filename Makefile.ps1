# Build the application
build:
	go build -o bin/volcanion-stress-test.exe cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go

# Run with debug logging
run-debug:
	$env:LOG_LEVEL="debug"; go run cmd/server/main.go

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Clean build artifacts
clean:
	Remove-Item -Recurse -Force bin -ErrorAction SilentlyContinue
	Remove-Item -Recurse -Force coverage.out -ErrorAction SilentlyContinue
	Remove-Item -Recurse -Force coverage.html -ErrorAction SilentlyContinue

# Run with custom port
run-port-9090:
	$env:SERVER_PORT="9090"; go run cmd/server/main.go

# Docker build (if Docker is available)
docker-build:
	docker build -t volcanion-stress-test:latest .

# Docker run
docker-run:
	docker run -p 8080:8080 volcanion-stress-test:latest

.PHONY: build run run-debug deps test test-coverage fmt lint clean run-port-9090 docker-build docker-run
