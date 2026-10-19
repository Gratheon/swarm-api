# Testing Documentation

## Test Organization

Tests are organized by type:
- **Unit Tests** (`*_unit_test.go`) - No external dependencies
- **Integration Tests** (`*_integration_test.go`) - Require database/services
- **E2E Tests** (`*_e2e_test.go`) - Full end-to-end scenarios

Integration tests are enabled with the `integration` build tag.

## Test Structure Standards

All tests follow these guidelines:

1. **Human-readable names** - Test names clearly describe what is being tested
2. **Nested structure** - Use `t.Run()` to organize test cases
3. **AAA pattern** - Include `// ARRANGE`, `// ACT`, `// ASSERT` comments
4. **Parallel execution** - Use `t.Parallel()` for independent tests

Example:
```go
func TestSplitHiveMutation(t *testing.T) {
    t.Parallel()

    t.Run("split hive with new queen", func(t *testing.T) {
        t.Parallel()

        t.Run("creates new hive with new queen and moves frames", func(t *testing.T) {
            t.Parallel()

            // ARRANGE
            db := setupTestDB(t)
            // ... setup code

            // ACT
            result, err := resolver.SplitHive(ctx, ...)

            // ASSERT
            if err != nil {
                t.Fatalf("SplitHive failed: %v", err)
            }
        })
    })
}
```

## Prerequisites

Integration tests require a running Postgres (18) database `swarm_api` (user/password `test`/`test`).

### 1. Start Postgres
Use the shared `postgres` dev service, or a standalone container:
```bash
docker run -d --name swarm-api-postgres -e POSTGRES_USER=test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=swarm_api -p 5432:5432 postgres:18
```

### 2. Run Migrations
```bash
cd ../swarm-api && just migrate-db-dev
```

### 3. Configure Database Connection (Optional)
Set environment variable:
```bash
export TEST_DB_DSN="postgres://test:test@localhost:5432/swarm_api?sslmode=disable"
```

## Running Tests

Run all tests:
```bash
go test -v ./...
```

Run specific test file:
```bash
cd graph && go test -v -tags=integration -run TestSplitHive
cd graph && go test -v -tags=integration -run TestDataLoader
```

Run specific test case:
```bash
cd graph && go test -v -tags=integration -run "TestSplitHiveMutation/split_hive_with_new_queen"
cd graph && go test -v -tags=integration -run "TestSplitHiveMutation/split_hive_by_taking_old_queen/moves_existing_queen"
```

Run tests with short timeout (skip slow integration tests):
```bash
go test -v -short ./...
```

## Test Coverage

### DataLoader Integration Tests (`dataloader_integration_test.go`)

Tests verify that DataLoaders batch concurrent queries correctly:

- **Hive Loader** - Batches queries for hives by apiary
- **Box Loader** - Batches queries for boxes by hive
- **Family Loader** - Batches queries for queens/families, handles missing families
- **Frame Loader** - Batches queries for frames by box
- **Frame Side Loader** - Batches queries for frame sides

Integration tests also verify DataLoaders work correctly with GraphQL resolvers.

### Split Hive Integration Tests (`split_hive_integration_test.go`)

Tests cover all split hive scenarios:

**Split with new queen:**
- Creates new hive with new queen
- Source hive keeps original queen
- Frames are moved correctly
- New queen has specified name

**Split by taking old queen:**
- Old queen is moved to new hive
- Source hive becomes queenless
- Returns error when source has no queen

**Split without queen:**
- New hive is queenless
- Source hive keeps its queen
- Frames are moved correctly

## Troubleshooting

If you're getting "source hive has no queen to take" errors:

1. Check if the queen has `hive_id` set in the families table
2. Run the migration if needed: `just migrate-db-dev`
