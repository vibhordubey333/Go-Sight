.PHONY: run compose-up compose-down compose-clean

run:
	go run .

compose-up:
	docker compose up -d --build

compose-down:
	docker compose down --remove-orphans

compose-clean:
	docker compose down --remove-orphans --volumes
