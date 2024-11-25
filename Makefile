# Version information
LAST_COMMIT := $(or $(shell git rev-parse --short HEAD 2> /dev/null),"unknown")
VERSION := $(or $(shell git describe --tags --abbrev=0 2> /dev/null),"v0.0.0")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%S%z")

# Build flags
LD_FLAGS := -s -w \
	-X 'main.version=${VERSION}' \
	-X 'main.commit=${LAST_COMMIT}' \
	-X 'main.date=${BUILD_DATE}'

# Tool configurations
GOPATH ?= $(shell go env GOPATH)
STUFFBIN ?= $(GOPATH)/bin/stuffbin
MOCKERY ?= $(GOPATH)/bin/mockery
PNPM ?= pnpm
GO ?= $(shell which go)

# Build configurations
BIN := inbox451
STATIC := frontend/dist:/

# Frontend paths and dependencies
FRONTEND_MODULES = frontend/node_modules
FRONTEND_DIST = frontend/dist
FRONTEND_DEPS = \
	$(FRONTEND_MODULES) \
	frontend/index.html \
	frontend/package.json \
	frontend/vite.config.ts \
	frontend/tsconfig.app.json \
	frontend/tsconfig.node.json \
	frontend/tailwind.config.js \
	frontend/components.json \
	$(shell find frontend/src frontend/public -type f)

.PHONY: build deps test dev pack-bin \
        build-frontend run-frontend \
        db-up db-down db-clean db-reset db-init db-install db-upgrade \
        release-dry-run release-snapshot release-tag install-goreleaser \
        fmt lint mocks

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

# Start development server
dev: db-up db-init
	CGO_ENABLED=0 $(GO) run -ldflags="${LD_FLAGS}" cmd/*.go

# Install all dependencies
deps: $(STUFFBIN)
	go mod download
	cd frontend && $(PNPM) install

# Run tests
test:
	go test -v ./...

# ==================================================================================== #
# TESTING & MOCKING
# ==================================================================================== #

# Install mockery
install-mockery:
	@echo "==> Installing mockery..."
	go get github.com/vektra/mockery/v2
	go install github.com/vektra/mockery/v2

# Generate mocks
mocks: install-mockery
	@echo "==> Generating mocks..."
	@$(MOCKERY)
	@echo "==> Mocks generated successfully"

# Run tests with coverage
test-coverage:
	@echo "==> Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

# Clean test cache and generated mocks
clean-test:
	@echo "==> Cleaning test cache and mocks..."
	@go clean -testcache
	@rm -rf internal/mocks coverage.out

# ==================================================================================== #
# FRONTEND
# ==================================================================================== #

# Install frontend dependencies
$(FRONTEND_MODULES): frontend/package.json frontend/pnpm-lock.yaml
	cd frontend && $(PNPM) install
	touch -c $(FRONTEND_MODULES)

# Build the frontend
$(FRONTEND_DIST): $(FRONTEND_DEPS)
	cd frontend && $(PNPM) build
	touch -c $(FRONTEND_DIST)

build-frontend: $(FRONTEND_DIST)

# Run frontend dev server
run-frontend:
	cd frontend && $(PNPM) dev

# ==================================================================================== #
# DATABASE
# ==================================================================================== #

# Start the database
db-up:
	docker compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3

# Stop the database
db-down:
	docker compose down

# Clean database volumes
db-clean:
	docker compose down -v
	docker volume rm inbox451_postgres_data || true
	docker rm inbox451-db-1 || true

# Reset database (down, clean, up)
db-reset: db-clean db-up db-init

# Initialize database
db-init:
	@CGO_ENABLED=0 $(GO) run -ldflags="${LD_FLAGS}" cmd/*.go --install --yes --idempotent || true

# Install database schema
db-install:
	CGO_ENABLED=0 $(GO) run -ldflags="${LD_FLAGS}" cmd/*.go --install --yes

# Upgrade database schema
db-upgrade:
	CGO_ENABLED=0 $(GO) run -ldflags="${LD_FLAGS}" cmd/*.go --upgrade --yes

# ==================================================================================== #
# BUILD
# ==================================================================================== #

# Install required tools
$(STUFFBIN):
	go install github.com/knadh/stuffbin/...

# Build the backend
build:
	CGO_ENABLED=0 $(GO) build -o ${BIN} -ldflags="${LD_FLAGS}" cmd/*.go

# Production build with embedded frontend
pack-bin: $(STUFFBIN) build build-frontend
	$(STUFFBIN) -a stuff -in ${BIN} -out ${BIN} ${STATIC}

# ==================================================================================== #
# RELEASE
# ==================================================================================== #

# Install goreleaser
install-goreleaser:
	go install github.com/goreleaser/goreleaser@latest

# Test the release process without publishing
release-dry-run: install-goreleaser
	goreleaser release --snapshot --clean --skip=publish

# Create a snapshot release for testing
release-snapshot: install-goreleaser
	goreleaser release --snapshot --clean

# Create and push a new release tag
release-tag:
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required. Use: make release-tag VERSION=v1.0.0"; exit 1; fi
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

# Format code
fmt:
	@echo "==> Formatting code..."
	go run mvdan.cc/gofumpt@latest -w .
	go run golang.org/x/tools/cmd/goimports@latest -w -local github.com/inbox451/inbox451 .

# Lint code
lint:
	@echo "==> Linting code..."
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run --fix
