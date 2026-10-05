.PHONY: up down logs test tidy sqlc web web-build web-test lint prod-build

COMPOSE := docker compose -f deploy/dev/docker-compose.yml

## Dev stack (db, mailpit, server w/ hot reload, web w/ Vite)
up:
	$(COMPOSE) up --build

down:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f server web

## Server
test:
	cd server && go test ./... -count=1

tidy:
	cd server && go mod tidy

sqlc:
	cd server && go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate

lint:
	cd server && go vet ./...
	cd web && npm run lint && npm run typecheck

## Web (outside compose)
web:
	cd web && npm install && npm run dev

web-build:
	cd web && npm run build

web-test:
	cd web && npm test

## Production images
prod-build:
	docker compose -f deploy/prod/docker-compose.yml build
