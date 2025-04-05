#!/bin/bash
set -e
cd /www/swarm-api/

echo "Stopping existing containers..."
COMPOSE_PROJECT_NAME=gratheon docker-compose down

# Build and migrations are now handled inside the Docker container via entrypoint.sh
echo "Starting new containers (build and migrations will run inside)..."
COMPOSE_PROJECT_NAME=gratheon docker-compose up -d --build
