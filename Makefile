# Get version info
LAST_COMMIT := $(or $(shell git rev-parse --short HEAD 2> /dev/null),"unknown")
VERSION := $(or $(shell git describe --tags --abbrev=0 2> /dev/null),"v0.0.0")
BUILDSTR := ${VERSION} (\#${LAST_COMMIT} $(shell date -u +"%Y-%m-%dT%H:%M:%S%z"))

# Tool paths
GOPATH ?= $(HOME)/go
STUFFBIN ?= $(GOPATH)/bin/stuffbin
PNPM ?= pnpm

# Frontend paths
FRONTEND_MODULES = frontend/node_modules
FRONTEND_DIST = frontend/dist
FRONTEND_DEPS = \
	$(FRONTEND_MODULES) \
	frontend/index.html \
	frontend/package.json \
	frontend/vite.config.js \
	$(shell find frontend/src frontend/public -type f)

BIN := inbox451
STATIC := frontend/dist:/

.PHONY: build deps build-frontend run-frontend dev test pack-bin

# Install required tools
$(STUFFBIN):
	go install github.com/knadh/stuffbin/...

# Frontend dependencies
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

# Build the backend
build:
	CGO_ENABLED=0 go build -o ${BIN} -ldflags="-s -w -X 'main.buildString=${BUILDSTR}' -X 'main.versionString=${VERSION}'" cmd/*.go

# Run the backend in dev mode
dev: db-up
	CGO_ENABLED=0 go run -ldflags="-s -w -X 'main.buildString=${BUILDSTR}' -X 'main.versionString=${VERSION}' -X 'main.frontendDir=frontend/dist'" cmd/*.go

# Database operations
db-up:
	docker compose up -d db

db-down:
	docker compose down

db-clean:
	docker compose down -v

db-reset: db-down db-clean db-up

# Testing
test:
	go test -v ./...

# Production build with embedded frontend
pack-bin: $(STUFFBIN) build build-frontend
	$(STUFFBIN) -a stuff -in ${BIN} -out ${BIN} ${STATIC}

# Install all dependencies
deps: $(STUFFBIN)
	go mod download
	cd frontend && $(PNPM) install
