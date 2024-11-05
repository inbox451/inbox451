# Mercury

A simple email server that allows you to create inboxes and rules to filter emails, written in Go.

## Table of Contents

- [Mercury](#mercury)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Prerequisites](#prerequisites)
  - [Development Setup](#development-setup)
  - [Configuration](#configuration)
    - [Using Config File](#using-config-file)
    - [Using Environment Variables](#using-environment-variables)
  - [Running the Server](#running-the-server)
  - [API Examples](#api-examples)
    - [Account Management](#account-management)
    - [Inbox Management](#inbox-management)
    - [Rule Management](#rule-management)
  - [Testing Email Reception](#testing-email-reception)
  - [Development](#development)
    - [Project Structure](#project-structure)
    - [API Testing with Bruno](#api-testing-with-bruno)
  - [Frontend Development](#frontend-development)
    - [Development Mode](#development-mode)
    - [Production Build](#production-build)
  - [Architecture](#architecture)

## Features

- HTTP API for managing accounts, inboxes, and rules
- SMTP server for receiving emails
- Rule-based email filtering
- Configurable via YAML and environment variables

## Prerequisites

- Go 1.22 or later
- Docker and Docker Compose
- Make (optional, but recommended)

## Development Setup

1. Clone the repository
2. Install dependencies:
```bash
make deps  # Installs Go and frontend dependencies
```

3. Start the development servers:

For backend development:
```bash
make dev  # Starts PostgreSQL and Go server
```

For frontend development:
```bash
make run-frontend  # Starts Vite dev server
```

Additional commands:
```bash
# Build the frontend
make build-frontend

# Build production binary (includes frontend)
make pack-bin

# Start the database
make db-up

# Stop the database
make db-down

# Clean database (remove volume)
make db-clean

# Reset database (down, clean, up)
make db-reset

# Run tests
make test
```

## Configuration

The application can be configured using:
- YAML configuration file (default: `config/default.yaml`)
- Environment variables
- Command line flags

### Using Config File

```yaml
server:
  http:
    port: ":8080"
  smtp:
    port: ":1025"
    hostname: "localhost"
    username: ""
    password: ""
  imap:
    port: ":1143"
    hostname: "localhost"
database:
  url: "postgres://mercury:mercury@localhost:5432/mercury?sslmode=disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m
logging:
  level: "info"  # Available: debug, info, warn, error, fatal
  format: "json"
```

### Using Environment Variables

Environment variables override config file settings:

```shell
MERCURY_SERVER_HTTP_PORT=":9090"
MERCURY_SERVER_SMTP_PORT=":2025"
MERCURY_DATABASE_URL="postgres://user:pass@host:5432/dbname"
MERCURY_LOGGING_LEVEL="debug"
```

## Running the Server

```shell
# Using make (recommended)
make dev

# Manual start
docker compose up -d db
go run cmd/mercury/main.go
```

## API Examples

### Account Management

Create an Account:
```shell
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Account"}'
```

List Accounts (with pagination):
```shell
curl "http://localhost:8080/accounts?limit=10&offset=0"
```

### Inbox Management

Create an Inbox:
```shell
curl -X POST http://localhost:8080/accounts/1/inboxes \
  -H "Content-Type: application/json" \
  -d '{"email": "inbox@example.com"}'
```

List Inboxes for Account (with pagination):
```shell
curl "http://localhost:8080/accounts/1/inboxes?limit=10&offset=0"
```

### Rule Management

Create a Rule:
```shell
curl -X POST http://localhost:8080/accounts/1/inboxes/1/rules \
  -H "Content-Type: application/json" \
  -d '{
    "sender": "sender@example.com",
    "receiver": "inbox@example.com",
    "subject": "Test Subject"
  }'
```

List Rules for Inbox:
```shell
curl http://localhost:8080/accounts/1/inboxes/1/rules
```

## Testing Email Reception

Using SWAKS:
```shell
swaks --to inbox@example.com \
      --from sender@example.com \
      --server localhost:1025 \
      --header "Subject: Test Subject" \
      --body "This is a test email."
```

Using Telnet:
```shell
telnet localhost 1025
HELO localhost
MAIL FROM:<sender@example.com>
RCPT TO:<inbox@example.com>
DATA
Subject: Test Subject

This is a test email.
.
QUIT
```

## Development

### Project Structure
```
.
├── Makefile               # Build and development commands
├── cmd/
│   └── main.go           # Main application entry point
├── config/
│   └── default.yaml      # Default configuration
├── frontend/             # Vue.js frontend application
│   ├── src/             # Frontend source code
│   │   ├── App.vue      # Root Vue component
│   │   ├── main.js      # Frontend entry point
│   │   └── style.css    # Global styles
│   ├── public/          # Static assets
│   ├── index.html       # HTML template
│   └── vite.config.js   # Vite configuration
├── internal/
│   ├── api/             # HTTP API implementation
│   ├── assets/          # Asset embedding (frontend)
│   ├── config/          # Configuration management
│   ├── core/            # Business logic
│   ├── logger/          # Logging package
│   ├── models/          # Data models
│   ├── smtp/            # SMTP server
│   ├── imap/           # IMAP server
│   └── storage/         # Database operations
```

### API Testing with Bruno

The project includes a comprehensive API test suite using [Bruno](https://www.usebruno.com/). Bruno is a fast and git-friendly API client that allows testing and validating API endpoints.

To use the Bruno collection:
1. Install Bruno from [usebruno.com](https://www.usebruno.com/)
2. Open the `bruno` folder in Bruno
3. Run requests individually or use collections
4. Tests will automatically validate responses

## Frontend Development

The project includes a Vue.js frontend that's served directly from the Go binary in production. In development, it runs on a separate Vite dev server.

### Development Mode

1. Start the backend:
```bash
make dev  # Runs on http://localhost:8080
```

2. In another terminal, start the frontend:
```bash
make run-frontend  # Runs on http://localhost:5173
```

### Production Build

The frontend is embedded into the Go binary using stuffbin:

```bash
make pack-bin
./inbox451 -config config/default.yaml
```

This creates a single binary that serves both the API and frontend assets. Visit http://localhost:8080 to access the application.

## Architecture

The application follows a layered architecture:
- API Layer: HTTP handlers and SMTP server
- Service Layer: Business logic in core package
- Storage Layer: Database operations

Database schema is inspired by [Archiveopteryx](https://archiveopteryx.org/db/)
