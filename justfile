start:
	make migrate-db-dev
	make build
	COMPOSE_PROJECT_NAME=gratheon docker compose -f docker-compose.dev.yml up --build

stop:
	COMPOSE_PROJECT_NAME=gratheon docker compose -f docker-compose.dev.yml down

develop:
	git rev-parse --short HEAD > .version
#	go run github.com/99designs/gqlgen generate
	NATIVE=1 ENV_ID=dev go run *.go

update:
	go get -u all

migrate-db-prod:
	go install github.com/pressly/goose/v3/cmd/goose@latest
	DSN=$$(jq -r '.db_dsn_migrate' config/config.prod.json) && $$(go env GOPATH)/bin/goose -dir migrations mysql "$$DSN" up

migrate-db-dev:
	go install github.com/pressly/goose/v3/cmd/goose@latest
	DSN=$$(jq -r '.db_dsn_migrate' config/config.dev.json) && $$(go env GOPATH)/bin/goose -dir migrations mysql "$$DSN" up

build:
	git rev-parse --short HEAD > .version
	@echo Building binary:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build \
		-a \
		-o swarm-api \
		*.go

gen:
	go get github.com/99designs/gqlgen
	go get github.com/99designs/gqlgen/internal/imports@v0.17.20
	go get github.com/99designs/gqlgen/codegen/config@v0.17.20
	go get -d
	@echo Generating schema.resolvers.go based on schema.graphql:
	go run github.com/99designs/gqlgen generate

