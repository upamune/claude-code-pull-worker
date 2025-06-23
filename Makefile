.PHONY: build run test clean install-service migrate

DB_FILE := claude-code-pull-worker.db
SCHEMA_FILE := sql/schema.sql

# Build the application
build:
	CGO_ENABLED=1 go build -o claude-code-pull-worker cmd/server/*.go

# Build for Linux AMD64
build-linux-amd64:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o claude-code-pull-worker-linux-amd64 cmd/server/*.go

# Run the application
run: migrate
	go run cmd/server/*.go

# Apply database migrations using sqlite3def
migrate:
	@if [ ! -f $(DB_FILE) ]; then \
		echo "Creating new database..."; \
		touch $(DB_FILE); \
	fi
	sqlite3def -f $(SCHEMA_FILE) $(DB_FILE)
	@echo "Applying seed data..."
	@sqlite3 $(DB_FILE) < sql/seed.sql || true

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f claude-code-pull-worker

# Install dependencies
deps:
	go mod download

# Generate systemd service file
systemd-gen:
	./claude-code-pull-worker systemd-install --user=$(USER) --working-dir=$(PWD)

# Install as systemd service (requires sudo)
install-service:
	@echo "Installing systemd service..."
	@echo "Please run: sudo make install-service-root"

install-service-root:
	cp claude-code-pull-worker.service /etc/systemd/system/
	systemctl daemon-reload
	systemctl enable claude-code-pull-worker
	@echo "Service installed. Start with: systemctl start claude-code-pull-worker"

# Development with auto-reload (requires air)
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Please install air first: go install github.com/cosmtrek/air@latest"; \
		exit 1; \
	fi

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "Please install golangci-lint first: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi
