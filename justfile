start:
    COMPOSE_PROJECT_NAME=gratheon docker compose -f docker-compose.dev.yml up --build

stop:
    COMPOSE_PROJECT_NAME=gratheon docker compose -f docker-compose.dev.yml down

develop:
    git rev-parse --short HEAD > .version
    # go run github.com/99designs/gqlgen generate
    NATIVE=1 ENV_ID=dev go run *.go

update:
    go get -u all

migrate-db-dev:
    go install github.com/pressly/goose/v3/cmd/goose@latest
    DSN=$(jq -r '.db_dsn_migrate' config/config.dev.json) && \
        echo $DSN && \
        $(go env GOPATH)/bin/goose -dir migrations mysql "$DSN" up

build:
    git rev-parse --short HEAD > .version
    @echo Building binary:
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
        -a \
        -o swarm-api \
        *.go

gen:
    # Ensure gqlgen tool is available
    go get github.com/99designs/gqlgen
    # Removed lines forcing specific old sub-package versions
    # Let go get resolve dependencies based on go.mod
    go get -d # Note: -d flag is deprecated but kept for compatibility if needed
    @echo Generating schema.resolvers.go based on schema.graphql:
    go run github.com/99designs/gqlgen generate
    @echo Updating version file for schema registry:
    git rev-parse --short HEAD > .version
    @echo "Schema generation complete! Version:" $(cat .version)
    @echo "⚠️  Remember to restart swarm-api to push the new schema to the registry"

test-specific:
    @echo "Running split hive e2e tests..."
    @echo "Make sure mysql is running: cd ../mysql && just start"
    cd graph && TESTING=true go test -v -run TestSplitHive

test-specific-verbose:
    @echo "Running split hive tests with verbose output..."
    cd graph && TESTING=true go test -v -run TestSplitHive -test.v

test:
    @echo "Running all tests..."
    TESTING=true go test -v ./...

