# Split Hive E2E Tests

## Prerequisites

Tests require a running MySQL database from the [mysql repo](https://github.com/Gratheon/mysql).

### 1. Start MySQL
```bash
cd ../mysql
just start
```

### 2. Configure Database Connection
Update `config/config.dev.json` with the mysql repo credentials:

```json
{
  "db_dsn": "root:test@tcp(localhost:5100)/swarm-api?parseTime=true"
}
```

Or set the environment variable:
```bash
export TEST_DB_DSN="root:test@tcp(localhost:5100)/swarm-api?parseTime=true"
```

### 3. Run Migrations
```bash
just migrate-db-dev
```

## Running Tests

Run all split hive tests:
```bash
just test-split
```

Or manually:
```bash
cd graph
go test -v -run TestSplitHive
```

Run a specific test:
```bash
cd graph
go test -v -run TestSplitHive_TakeOldQueen
```

## Test Coverage

### TestSplitHive_NewQueen
Tests creating a split with a new queen:
- Source hive keeps its original queen
- New hive gets a queen with the specified name
- Frames are moved correctly

### TestSplitHive_TakeOldQueen  
Tests moving the old queen to the new hive:
- Old queen is moved to new hive
- Source hive becomes queenless
- Both `hives.family_id` and `families.hive_id` are updated correctly

### TestSplitHive_NoQueen
Tests creating a queenless split:
- Source hive keeps its queen
- New hive has no queen
- Frames are moved correctly

### TestSplitHive_TakeOldQueen_NoQueenInSource
Tests error handling:
- Attempting to take queen from queenless hive fails with appropriate error

### TestSplitHive_LegacyQueenTracking
Tests backward compatibility:
- Handles queens tracked only via `hives.family_id` (legacy)
- Properly migrates to `families.hive_id` tracking
- Cleans up `hives.family_id` reference

## Debug Query

Use the `debugHiveQueens` query to inspect queen relationships:

```graphql
query {
  debugHiveQueens(hiveId: "123")
}
```

This returns detailed information about:
- `hives.family_id` value
- Queen name via legacy relationship
- `families.hive_id` value for the queen
- Queens found via `families.hive_id` 
- Direct database query count

## Troubleshooting

If you're getting "source hive has no queen to take" errors:

1. Run the debug query to check both relationship types
2. Check if the queen has `hive_id` set in the families table
3. Check if the hive has `family_id` set in the hives table
4. Run the migration if needed: `just migrate-db-dev`

