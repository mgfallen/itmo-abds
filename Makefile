.PHONY: run stop gen load setup teardown

run:
	docker compose --env-file .env.local up -d --build

stop:
	docker compose --env-file .env.local down

gen:
	go run cmd/gen/main.go

load:
	go run cmd/load/main.go

setup: run
	sleep 5
	go run cmd/gen/main.go
	go run cmd/load/main.go

teardown:
	docker compose --env-file .env.local down -v