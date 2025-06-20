<p align="center">
  <img src="frontend/public/logo.png" alt="Inbox451 Logo" width="200"/>
</p>

# Inbox451

A simple email server that allows you to create inboxes and rules to filter emails, written in Go.

[![Go Report Card](https://goreportcard.com/badge/github.com/inbox451/inbox451)](https://goreportcard.com/report/github.com/inbox451/inbox451)
[![Build Status](https://github.com/inbox451/inbox451/actions/workflows/pull-request.yml/badge.svg)](https://github.com/inbox451/inbox451/actions/workflows/pull-request.yml)
[![codecov](https://codecov.io/gh/inbox451/inbox451/graph/badge.svg?token=4HPWU0V3YD)](https://codecov.io/gh/inbox451/inbox451)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Features

- HTTP API for managing projects, inboxes, and rules
- SMTP server for receiving emails
- IMAP server for accessing emails
- Rule-based email filtering
- Configurable via YAML and environment variables

## Quick Start

1. Install dependencies:
```bash
make deps
```

2. Start the development servers:
```bash
make dev          # Starts PostgreSQL and Go server
make run-frontend # In another terminal, starts Vite dev server
```

Visit http://localhost:8080 for the API and http://localhost:5173 for the development frontend.

## Configuration

The application can be configured via:
1. Environment variables (highest precedence)
2. Configuration file
3. Default values

Example configuration:
```yaml
server:
  http:
    port: ":8080"
  smtp:
    domain: "smtp.example.com"
    hostname: "localhost"
    allow_insecure_auth: true
    msa:
      port: "587"
      tls: false
    mta:
      port: "1025"
      tls: false
  imap:
    port: ":1143"
    hostname: "localhost"
    tls: false
    allow_insecure_auth: true
  # Shared TLS certificate configuration for all services
  tls:
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"
  email_domain: "example.com"
database:
  url: "postgres://inbox:inbox@localhost:5432/inbox451?sslmode=disable"
logging:
  level: "info"
  format: "json"
```

### TLS Configuration

To enable TLS/STARTTLS for SMTP and IMAP servers:

1. Generate or obtain SSL/TLS certificates
2. Update the configuration with certificate paths in the `server.tls` section
3. Enable TLS for specific services:
   - Set `server.smtp.msa.tls: true` for SMTP MSA (port 587)
   - Set `server.smtp.mta.tls: true` for SMTP MTA (port 25)
   - Set `server.imap.tls: true` for IMAP (port 1143)
4. Set `allow_insecure_auth: false` to require encrypted connections for authentication

Example TLS-enabled configuration:
```yaml
server:
  smtp:
    allow_insecure_auth: false
    msa:
      tls: true
  imap:
    tls: true
    allow_insecure_auth: false
  tls:
    cert_file: "/etc/ssl/certs/mail.example.com.crt"
    key_file: "/etc/ssl/private/mail.example.com.key"
```

## API Examples

Create a Project:
```shell
curl -X POST http://localhost:8080/api/projects \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Project"}'
```

Create an Inbox:
```shell
curl -X POST http://localhost:8080/api/projects/1/inboxes \
  -H "Content-Type: application/json" \
  -d '{"email": "inbox@example.com"}'
```

Create a Rule:
```shell
curl -X POST http://localhost:8080/api/projects/1/inboxes/1/rules \
  -H "Content-Type: application/json" \
  -d '{
    "sender": "sender@example.com",
    "receiver": "inbox@example.com",
    "subject": "Test Subject"
  }'
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

## Project Structure
```
.
├── cmd/                # Application entry points
├── frontend/           # Vue.js frontend application
├── internal/           # Internal packages
│   ├── api/            # HTTP API implementation
│   ├── core/           # Business logic
│   ├── smtp/           # SMTP server
│   ├── imap/           # IMAP server
│   ├── migrations/     # Database migrations
│   ├── storage/        # Database repositories
│   └── models/         # Database models
└── bruno/              # API test collections
```

## Documentation

- [Contributing Guide](CONTRIBUTING.md)
- [Release Process](RELEASE.md)
- [API Documentation](docs/api.md) (**WIP**)

## License

[MIT](LICENSE)
