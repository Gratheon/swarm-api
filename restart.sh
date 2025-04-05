#!/bin/bash
set -e

cd /www/swarm-api/
make build
COMPOSE_PROJECT_NAME=gratheon docker-compose down

echo "Running database migrations..."
go install github.com/pressly/goose/v3/cmd/goose@latest
DSN=$(jq -r '.db_dsn_migrate' config/config.prod.json) && $(go env GOPATH)/bin/goose -dir migrations mysql "$DSN" up
echo "Migrations complete."

COMPOSE_PROJECT_NAME=gratheon docker-compose up -d --build
