.PHONY: docs test tidy up app seed

API := api

docs:
	cd $(API) && go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g main.go -d ./cmd,./routes,./services -o ./docs

test:
	cd $(API) && go test ./... -count=1

tidy:
	cd $(API) && go mod tidy

up:
	docker compose -f infra/dev/docker-compose.yml up --build

app:
	cd mobile && npx expo start --web

seed:
	bash infra/dev/seed.sh
