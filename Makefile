start:
	make build && COMPOSE_PROJECT_NAME=gratheon docker compose -f docker-compose.dev.yml up --build
develop:
	git rev-parse --short HEAD > .version
#	go run github.com/99designs/gqlgen generate
	NATIVE=1 ENV_ID=dev go run *.go

update:
	go get -u all

build:
#	git rev-parse --short HEAD > .version
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

deploy-copy:
	scp -r Dockerfile schema.graphql .version docker-compose.yml restart.sh root@gratheon.com:/www/api.gratheon.com/
	scp -r ./swarm-api root@gratheon.com:/www/api.gratheon.com/
	scp -r config/* root@gratheon.com:/www/api.gratheon.com/config/

deploy-run:
	ssh root@gratheon.com 'bash /www/api.gratheon.com/restart.sh'

deploy:
	git rev-parse --short HEAD > .version
	make build
	make deploy-copy
	make deploy-run

.PHONY: run
