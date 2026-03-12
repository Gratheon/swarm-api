# ---- Build Stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Install build tools
RUN apk add --no-cache git

# Copy Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Refresh .version from current git commit before compilation.
RUN ./scripts/update-version.sh

# Build the swarm-api application
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o swarm-api *.go

# Build goose with only MySQL support to keep the runtime binary small.
RUN GOBIN=/build/bin CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go install -trimpath -ldflags="-s -w" \
    -tags="no_clickhouse no_libsql no_mssql no_postgres no_sqlite3 no_vertica no_ydb" \
    github.com/pressly/goose/v3/cmd/goose@v3.27.0

# ---- Final Stage ----
FROM alpine:3.20

WORKDIR /app

# Create a non-root user and switch to it
RUN apk add --no-cache jq && addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy runtime files with final ownership and permissions.
COPY --from=builder --chown=appuser:appgroup /build/swarm-api /app/swarm-api
COPY --from=builder --chown=appuser:appgroup /build/bin/goose /app/goose
COPY --chown=appuser:appgroup migrations /app/migrations
COPY --chown=appuser:appgroup config /app/config
COPY --chown=appuser:appgroup entrypoint.sh /app/entrypoint.sh

USER appuser

EXPOSE 8100

# Set the entrypoint script
ENTRYPOINT ["/app/entrypoint.sh"]
