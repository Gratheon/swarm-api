# ---- Build Stage ----
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install build tools
RUN apk add --no-cache git

# Copy Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the swarm-api application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o swarm-api *.go

# Build the goose migration tool
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# ---- Final Stage ----
FROM alpine:3.18

WORKDIR /app

# Install runtime dependencies (jq for entrypoint script)
RUN apk add --no-cache jq

# Copy the built application binary from the builder stage
COPY --from=builder /build/swarm-api /app/swarm-api

# Copy the built goose binary from the builder stage
COPY --from=builder /go/bin/goose /app/goose

# Copy migrations, config, and entrypoint script
COPY migrations /app/migrations
COPY config /app/config
COPY entrypoint.sh /app/entrypoint.sh

# Ensure scripts and binaries are executable
RUN chmod +x /app/swarm-api /app/goose /app/entrypoint.sh

# Create a non-root user and switch to it
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8100

# Set the entrypoint script
ENTRYPOINT ["/app/entrypoint.sh"]
