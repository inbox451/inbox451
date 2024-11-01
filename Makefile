.PHONY: dev db-up db-down db-clean test

dev: db-up
	go run ./cmd/mercury

db-up:
	docker compose up -d db

db-down:
	docker compose down

# Remove the database volume completely
db-clean:
	docker compose down -v

# Reset database (down, clean volumes, up)
db-reset: db-down db-clean db-up

test:
	go test -v ./...