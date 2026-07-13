#!/bin/sh
set -eu

PROJECT_NAME="${COMPOSE_PROJECT_NAME:-swarm-api}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.yml}"
LEGACY_PROJECT_NAME="${LEGACY_PROJECT_NAME:-gratheon}"
SERVICE_NAME="${SERVICE_NAME:-swarm-api}"

cd "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"

stop_legacy_gratheon_service_container() {
  docker ps -aq \
    --filter "label=com.docker.compose.project=${LEGACY_PROJECT_NAME}" \
    --filter "label=com.docker.compose.service=${SERVICE_NAME}" \
    | while IFS= read -r container_id; do
        [ -n "$container_id" ] || continue
        docker rm -f "$container_id" >/dev/null 2>&1 || true
      done
}

echo "Stopping existing containers..."
stop_legacy_gratheon_service_container
COMPOSE_PROJECT_NAME="$PROJECT_NAME" docker-compose -f "$COMPOSE_FILE" down

echo "Starting new containers (build and migrations will run inside)..."
VERSION=$(git rev-parse --short HEAD)
export VERSION
COMPOSE_PROJECT_NAME="$PROJECT_NAME" docker-compose -f "$COMPOSE_FILE" up -d --build
