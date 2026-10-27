# gratheon / swarm-api
[![codecov](https://codecov.io/gh/Gratheon/swarm-api/graph/badge.svg?token=ZOJQ9HEHU3)](https://codecov.io/gh/Gratheon/swarm-api)

Main monolith service to manage beehive data.


## Architecture

```mermaid
flowchart LR
    web-app("<a href='https://github.com/Gratheon/web-app'>web-app</a>") --> graphql-router
    web-app --"subscribe to events"--> event-stream-filter("<a href='https://github.com/Gratheon/event-stream-filter'>event-stream-filter</a>") --> redis
    
    graphql-router --> swarm-api("<a href='https://github.com/Gratheon/swarm-api'>swarm-api</a>") --> postgres[(postgres)]
    graphql-router --> swarm-api --> redis[("<a href='https://github.com/Gratheon/redis'>redis pub-sub</a>")]
    
    graphql-router --> graphql-schema-registry
```

## Database migrations
We use goose to manage database migrations.
To 
```bash
just migrate-db-dev
```


## Development
Based on [gqlgen](https://gqlgen.com/getting-started/).

```
cp config/config.example.json config/config.dev.json
just start
just setup-git-hooks
```

To re-generate graphql -> go resolvers, run: 
```
just gen
```

When `schema.graphql` is changed, the pre-commit hook refreshes `.version` automatically.

## Testing

### Test Organization

Tests are organized by type using file suffixes:
- `*_unit_test.go` - Unit tests that don't require external dependencies
- `*_integration_test.go` - Integration tests that require database/external services
- `*_e2e_test.go` - End-to-end tests

Integration tests are additionally guarded by the `integration` build tag.

### Test Structure Guidelines

- Tests use human-readable names that clearly describe what is being tested
- Tests use `t.Run()` nesting to organize test cases and scenarios
- Each test includes `// ARRANGE`, `// ACT`, and `// ASSERT` comments for clarity
- Integration tests use `t.Parallel()` to run concurrently where possible
- Assertions use `github.com/stretchr/testify/assert` for type-safe, readable checks

### Running Tests

Run all tests:
```bash
go test -v ./...
```

Run specific test type:
```bash
go test -v ./...
go test -v -tags=integration ./graph
```

Run tests for a specific function:
```bash
cd graph && go test -v -tags=integration -run "TestSplitHiveMutation/split_hive_with_new_queen"
```

### Prerequisites

Integration tests require a running Postgres (18) database `swarm_api` (user/password `test`/`test`),
for example the shared `postgres` dev service or:
```bash
docker run -d --name swarm-api-postgres -e POSTGRES_USER=test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=swarm_api -p 5432:5432 postgres:18
cd ../swarm-api && just migrate-db-dev
```

Configure database connection (optional, this is the default):
```bash
export TEST_DB_DSN="postgres://test:test@localhost:5432/swarm_api?sslmode=disable"
```

See [graph/TESTING.md](graph/TESTING.md) for detailed testing documentation.

## Database connections

The production Postgres is a DigitalOcean managed cluster shared by several services, so
`max_connections` is small (25 on the smallest plan, 3 of them reserved for superusers).
swarm-api keeps one shared pool for the whole process:

- `db_max_open_conns` (default 8) caps open connections. Parallel GraphQL requests beyond
  that wait for a free connection instead of opening new ones.
- `db_max_idle_conns` (default = `db_max_open_conns`) keeps connections warm between requests.
- Opening a connection is retried for ~1s on `SQLSTATE 53300` (no free slots).
- Pool usage is exported on `/metrics` as `go_sql_*{db_name="swarm_api"}`; a growing
  `go_sql_wait_count_total` means requests are queuing for a connection.

Queries use pgx's `exec` mode (no named prepared statements), so `db_dsn` can point either
at the cluster directly or at a DigitalOcean connection pool (PgBouncer, transaction mode,
port 25061). Migrations (`db_dsn_migrate`) should keep connecting directly (port 25060).

Database and network errors are logged server-side and returned to GraphQL clients only as
`Database is temporarily unavailable, please retry` / `Internal server error`.

## Building
```
just build
```

## License

This project is **dual-licensed**:

- **[AGPL v3](LICENSE-AGPL.md)** - For open source use:
    - ✅ You can use, modify, and distribute the software
    - ✅ Source code is freely available
    - ⚠️ If you modify and host a public instance, you must share your modifications
    - ⚠️ Any derivative work must also be licensed under AGPL v3

- **[Commercial License](LICENSE-COMMERCIAL.md)** - For enterprise customers who need:
    - 🏢 On-premise deployment without source disclosure
    - 🔧 Custom modifications without copyleft obligations
    - 🎨 White-label/rebranding rights
    - 📞 Priority support and SLA guarantees

### Contributing

If you'd like to contribute, please see our [Contributing Guide](CONTRIBUTING.md) and sign our Contributor License Agreement (CLA).

*   Please refer to the `CODEOWNERS` file for ownership details.
*   Follow standard GitHub flow: Fork the repository, create a feature branch, make changes, and submit a Pull Request.
