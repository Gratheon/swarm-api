#!/bin/sh
set -e

# Default to production config if ENV_ID is not set or empty
CONFIG_FILE="config/config.${ENV_ID:-prod}.json"

echo "Using config file: $CONFIG_FILE"

# Check if config file exists
if [ ! -f "$CONFIG_FILE" ]; then
  echo "Error: Configuration file $CONFIG_FILE not found!"
  exit 1
fi

echo "Running database migrations..."
# Extract DSN using jq (ensure jq is installed in the final image)
DSN=$(jq -r '.db_dsn_migrate' "$CONFIG_FILE")

if [ -z "$DSN" ] || [ "$DSN" = "null" ]; then
  echo "Error: Could not read db_dsn_migrate from $CONFIG_FILE"
  exit 1
fi

# Run migrations using the goose binary copied from the build stage
/app/goose -dir migrations mysql "$DSN" up
echo "Migrations complete."

echo "Starting swarm-api..."
# Execute the main application binary
exec /app/swarm-api
