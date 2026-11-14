.PHONY: run stop gen load mock setup teardown migrate test test-integration

run:
	docker compose --env-file .env.local up -d --build

stop:
	docker compose --env-file .env.local down

gen:
	go run cmd/gen/main.go

load:
	go run cmd/load/main.go

mock:
	go generate ./internal/repository

migrate:
	docker compose --env-file .env.local exec -T clickhouse clickhouse-client --multiquery < internal/db/schema.sql

setup: run
	@echo "Waiting for ClickHouse..."
	@powershell -Command "while ($true) { try { docker compose --env-file .env.local exec -T clickhouse clickhouse-client --query 'SELECT 1' 2>&1 | Out-Null; if ($LASTEXITCODE -eq 0) { break } } catch {}; Start-Sleep -Seconds 1 }" || \
	sh -c 'until docker compose --env-file .env.local exec -T clickhouse clickhouse-client --query "SELECT 1" >/dev/null 2>&1; do sleep 1; done'
	docker compose --env-file .env.local exec -T clickhouse clickhouse-client --multiquery < internal/db/schema.sql
	go run cmd/gen/main.go
	go run cmd/load/main.go

test:
	go test -v ./internal/...

test-integration:
	go test -v -tags=integration ./internal/repository/...

teardown:
	docker compose --env-file .env.local down -v