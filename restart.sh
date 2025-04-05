#!/bin/bash
set -e
cd /www/swarm-api/

echo "Starting build process..."
git rev-parse --short HEAD > .version
@echo Building binary:
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
go build \
    -a \
    -o swarm-api \
    *.go

    
echo "Build complete."
echo "Stopping existing containers..."
COMPOSE_PROJECT_NAME=gratheon docker-compose down

echo "Running database migrations..."
go install github.com/pressly/goose/v3/cmd/goose@latest

echo "Running migrations..."
DSN=$(jq -r '.db_dsn_migrate' config/config.prod.json) && $(go env GOPATH)/bin/goose -dir migrations mysql "$DSN" up
echo "Migrations complete."

echo "Starting new containers..."
COMPOSE_PROJECT_NAME=gratheon docker-compose up -d --build
