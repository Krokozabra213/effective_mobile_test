
PG_DSN=postgres://myuser:mypassword@localhost:5555/postgres?sslmode=disable
MIGR_DIR=migrations/goose

.PHONY: migrate-create migrate-up migrate-down migrate-status migrate-reset test-integrate

# Default target
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
# 	@echo "  docker-up                            Start all containers"
# 	@echo "  docker-down                          Stop all containers"
	@echo "  migrate-create name=<table_name>     Create new migration file"
	@echo "  migrate-up                           Apply all pending migrations"
	@echo "  migrate-down                         Rollback last migration"
	@echo "  migrate-status                       Show migrations status"
	@echo "  migrate-reset                        Rollback all migrations"
	@echo "  tests                                Start tests"
	@echo "  test-integrate                       Start intergration tests"

# Start containers
# docker-up:
# 	docker compose up -d

# Stop containers
# docker-down:
# 	docker compose down

test-integrate:
	go test -tags=integration -v ./...

# Create new migration file: make migrate-create name=name_table
migrate-create:
	goose -dir $(MIGR_DIR) create $(name) sql

migrate-up:
	goose -dir $(MIGR_DIR) postgres "$(PG_DSN)" up

migrate-down:
	goose -dir $(MIGR_DIR) postgres "$(PG_DSN)" down

migrate-status:
	goose -dir $(MIGR_DIR) postgres "$(PG_DSN)" status

migrate-reset:
	goose -dir $(MIGR_DIR) postgres "$(PG_DSN)" reset

# Start tests
# tests:
# 	go test -v -count=1 ./tests/...
