.PHONY: build run test fmt vet clean

# Build the application
build:
	go build -o bin/bot ./cmd/bot

# Run the application
run:
	go run ./cmd/bot

# Run the worker
run-worker:
	go run ./cmd/worker	

# Run tests
test:
	go test ./...

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Run all checks
check: fmt vet test

# Install dependencies
deps:
	go mod download

tidy:
	go mod tidy