up:
	docker compose up

up-d:
	docker compose up -d

build:
	docker compose up --build

down:
	docker compose down

logs:
	docker compose logs -f

.PHONY: up up-d build down logs