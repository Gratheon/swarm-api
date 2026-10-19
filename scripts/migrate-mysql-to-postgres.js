#!/usr/bin/env node
/*
 * One-off copy of swarm-api data from MySQL to Postgres.
 *
 * WHY: the service switched its storage from MySQL to Postgres. Data volume is small
 * (~70k rows), so a batched read / batch-insert copy is simpler and more predictable than
 * pgloader (which would create its own schema and types).
 *
 * Prerequisite: the Postgres schema is created by goose (migrations/*_postgres_baseline.sql),
 * exactly like the service does on boot. This script only copies data.
 *
 * WHAT it does, in ONE transaction:
 *   1. truncates all tables (including the seed rows of the baseline migration)
 *   2. copies every row with its original id, in foreign key order and primary-key order,
 *      in batches; box_systems.box_profile_source_system_id references the same table, so it
 *      is filled in after the whole table is copied
 *   3. verifies per table: row counts and per-column content (non-null counts, sums of
 *      integers, text lengths, timestamps as epoch seconds, JSON element counts, booleans)
 *   4. resets identity sequences (skipped in DRY_RUN: setval is not transactional)
 * and rolls back on any mismatch. No data cleanup is done.
 *
 * MySQL is only read, so it stays intact as a rollback target.
 * All connection settings come from env variables; nothing environment specific lives here.
 *
 * Run (from the repo root, deps are installed into a throwaway dir):
 *   docker run --rm --network host -v "$PWD":/repo -w /repo \
 *     -e MYSQL_HOST -e MYSQL_PORT -e MYSQL_USER -e MYSQL_PASSWORD -e MYSQL_DATABASE \
 *     -e PG_HOST -e PG_PORT -e PG_USER -e PG_PASSWORD -e PG_DATABASE -e PG_SSL -e DRY_RUN -e FORCE \
 *     node:24-slim sh -c 'npm i --prefix /tmp/deps mysql2@3 pg@8 >/dev/null && NODE_PATH=/tmp/deps/node_modules node scripts/migrate-mysql-to-postgres.js'
 *
 * DRY_RUN=1 performs the full copy and verification, then rolls everything back.
 * FORCE=1 is required when Postgres already contains user data (e.g. a second rehearsal),
 * so that a re-run after cutover cannot silently wipe data written by the new service.
 */
const mysql = require("mysql2/promise");
const { Client } = require("pg");

// Foreign key order: parents before children
const TABLES = [
  "apiaries",
  "hives",
  "apiary_obstacles",
  "box_systems",
  "box_specs",
  "box_spec_frame_sources",
  "frame_specs",
  "frame_spec_compatibility",
  "boxes",
  "frames_sides",
  "frames",
  "devices",
  "families",
  "family_moves",
  "files",
  "hive_logs",
  "hive_placements",
  "inspections",
  "treatments",
  "warehouse_frame_inventory",
  "warehouse_modules",
  "warehouse_settings",
];
// Tables that the baseline migration seeds; they are not "user data" for the FORCE guard
const SEEDED_TABLES = new Set(["box_systems", "box_specs", "frame_specs", "frame_spec_compatibility", "box_spec_frame_sources"]);
// Self-referencing foreign keys: copied as NULL first, set after the whole table is in place
const SELF_REFERENCES = { box_systems: ["box_profile_source_system_id"] };

const BATCH_SIZE = 1000;
const DRY_RUN = process.env.DRY_RUN === "1";
const FORCE = process.env.FORCE === "1";

function requireEnv(name) {
  const value = process.env[name];
  if (!value) {
    throw new Error(`Missing env ${name}`);
  }
  return value;
}

// Column kinds drive both value conversion and verification
async function pgColumnKinds(pg, table) {
  const { rows } = await pg.query(
    "SELECT column_name, data_type FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = $1 ORDER BY ordinal_position",
    [table]
  );
  const kinds = {};
  for (const { column_name: name, data_type: type } of rows) {
    if (type === "boolean") kinds[name] = "bool";
    else if (type === "jsonb" || type === "json") kinds[name] = "json";
    else if (type.startsWith("timestamp")) kinds[name] = "time";
    else if (["integer", "bigint", "smallint"].includes(type)) kinds[name] = "int";
    else if (["double precision", "real", "numeric"].includes(type)) kinds[name] = "float";
    else kinds[name] = "text";
  }
  return kinds;
}

async function mysqlColumns(my, table) {
  const [rows] = await my.query(
    "SELECT COLUMN_NAME AS name FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION",
    [table]
  );
  return rows.map((r) => r.name);
}

function convertValue(kind, value) {
  if (value === null || value === undefined) {
    return null;
  }
  if (kind === "json") {
    // mysql2 parses JSON columns; re-serialize so node-pg does not turn arrays into PG array literals
    return JSON.stringify(value);
  }
  if (kind === "bool") {
    return Number(value) !== 0;
  }
  // MySQL zero dates have no Postgres equivalent (none on prod, guarded anyway)
  if (kind === "time" && typeof value === "string" && value.startsWith("0000-00-00")) {
    return null;
  }
  return value;
}

async function copyTable(my, pg, table, kinds) {
  const sourceColumns = await mysqlColumns(my, table);
  // Fail loudly instead of silently dropping data if schemas drifted
  const missing = sourceColumns.filter((c) => !(c in kinds));
  if (missing.length > 0) {
    throw new Error(`${table}: columns missing in Postgres: ${missing.join(", ")}`);
  }
  const deferred = SELF_REFERENCES[table] || [];

  const selectList = sourceColumns.map((c) => `\`${c}\``).join(", ");
  const insertList = sourceColumns.map((c) => `"${c}"`).join(", ");

  let copied = 0;
  let lastId = null;
  const deferredValues = [];
  for (;;) {
    // Keyset pagination on the primary key (every table has an "id" PK)
    const [batch] = await my.query(
      `SELECT ${selectList} FROM \`${table}\` ${lastId === null ? "" : "WHERE id > ?"} ORDER BY id LIMIT ${BATCH_SIZE}`,
      lastId === null ? [] : [lastId]
    );
    if (batch.length === 0) {
      break;
    }

    const params = [];
    const tuples = batch.map((row) => {
      const placeholders = sourceColumns.map((column) => {
        if (deferred.includes(column)) {
          if (row[column] !== null) deferredValues.push({ id: row.id, column, value: row[column] });
          params.push(null);
        } else {
          params.push(convertValue(kinds[column], row[column]));
        }
        return `$${params.length}`;
      });
      return `(${placeholders.join(", ")})`;
    });
    await pg.query(`INSERT INTO "${table}" (${insertList}) VALUES ${tuples.join(", ")}`, params);

    copied += batch.length;
    lastId = batch[batch.length - 1].id;
  }

  if (deferredValues.length > 0) {
    // The updated_at trigger must not touch copied rows (ALTER TABLE is transactional)
    await pg.query(`ALTER TABLE "${table}" DISABLE TRIGGER USER`);
    for (const { id, column, value } of deferredValues) {
      await pg.query(`UPDATE "${table}" SET "${column}" = $1 WHERE id = $2`, [value, id]);
    }
    await pg.query(`ALTER TABLE "${table}" ENABLE TRIGGER USER`);
  }

  return { columns: sourceColumns, copied };
}

// Compares per-column aggregates, not only row counts
async function verifyContent(my, pg, table, columns, kinds) {
  const checks = [];
  for (const c of columns) {
    const kind = kinds[c];
    checks.push([`COUNT(\`${c}\`)`, `COUNT("${c}")`]);
    if (kind === "int") {
      checks.push([`COALESCE(SUM(\`${c}\`), 0)`, `COALESCE(SUM("${c}"), 0)`]);
    } else if (kind === "text") {
      checks.push([`COALESCE(SUM(CHAR_LENGTH(\`${c}\`)), 0)`, `COALESCE(SUM(char_length("${c}")), 0)`]);
    } else if (kind === "time") {
      // MySQL zone is verified to be UTC, so UNIX_TIMESTAMP() and EXTRACT(EPOCH) must agree
      checks.push([`COALESCE(SUM(UNIX_TIMESTAMP(\`${c}\`)), 0)`, `COALESCE(SUM(EXTRACT(EPOCH FROM "${c}")), 0)`]);
    } else if (kind === "bool") {
      checks.push([`COALESCE(SUM(\`${c}\` <> 0), 0)`, `COUNT(*) FILTER (WHERE "${c}")`]);
    } else if (kind === "json") {
      // same semantics as MySQL JSON_LENGTH(): NULL for SQL NULL, 1 for a scalar
      checks.push([
        `COALESCE(SUM(JSON_LENGTH(\`${c}\`)), 0)`,
        `COALESCE(SUM(CASE
          WHEN "${c}" IS NULL THEN NULL
          WHEN jsonb_typeof("${c}") = 'array' THEN jsonb_array_length("${c}")
          WHEN jsonb_typeof("${c}") = 'object' THEN (SELECT COUNT(*) FROM jsonb_object_keys("${c}"))
          ELSE 1 END), 0)`,
      ]);
    } else if (kind === "float") {
      checks.push([`COALESCE(SUM(\`${c}\`), 0)`, `COALESCE(SUM("${c}"), 0)`, "float"]);
    }
  }
  const [[myRow]] = await my.query({ sql: `SELECT ${checks.map((x) => x[0]).join(", ")} FROM \`${table}\``, rowsAsArray: true });
  const { rows } = await pg.query({ text: `SELECT ${checks.map((x) => x[1]).join(", ")} FROM "${table}"`, rowMode: "array" });
  const pgRow = rows[0];
  const diffs = [];
  checks.forEach(([expr, , mode], i) => {
    const a = Number(myRow[i]);
    const b = Number(pgRow[i]);
    // FLOAT is 32-bit in MySQL and read via its text form, so allow rounding noise
    const equal = mode === "float" ? Math.abs(a - b) <= Math.max(1e-3, Math.abs(a) * 1e-5) : a === b;
    if (!equal) {
      diffs.push(`${expr}: mysql=${myRow[i]} pg=${pgRow[i]}`);
    }
  });
  if (diffs.length > 0) {
    throw new Error(`${table}: content mismatch: ${diffs.join("; ")}`);
  }
}

async function main() {
  const my = await mysql.createConnection({
    host: requireEnv("MYSQL_HOST"),
    port: Number(process.env.MYSQL_PORT || 3306),
    user: requireEnv("MYSQL_USER"),
    password: requireEnv("MYSQL_PASSWORD"),
    database: requireEnv("MYSQL_DATABASE"),
    // Keep DATETIME values as raw strings: no JS timezone conversion on the way through
    dateStrings: true,
    timezone: "Z",
    supportBigNumbers: true,
    charset: "utf8mb4",
  });

  const pg = new Client({
    host: requireEnv("PG_HOST"),
    port: Number(process.env.PG_PORT || 5432),
    user: requireEnv("PG_USER"),
    password: requireEnv("PG_PASSWORD"),
    database: requireEnv("PG_DATABASE"),
    // Managed Postgres needs TLS; its CA is not in the default trust store
    ssl: process.env.PG_SSL === "disable" ? false : { rejectUnauthorized: false },
  });
  await pg.connect();

  try {
    // DATETIME has no zone: reading it as UTC is only correct if the MySQL zone is UTC.
    // Abort instead of shifting data.
    const [[tz]] = await my.query("SELECT @@session.time_zone AS session_tz, @@system_time_zone AS system_tz");
    console.log(`MySQL time zone: session=${tz.session_tz} system=${tz.system_tz}`);
    const isUtc = (zone) => ["UTC", "+00:00", "Etc/UTC"].includes(zone);
    if (!(isUtc(tz.session_tz) || (tz.session_tz === "SYSTEM" && isUtc(tz.system_tz)))) {
      throw new Error("MySQL session time zone is not UTC; date values would be shifted");
    }
    await pg.query("SET TIME ZONE 'UTC'");

    const { rows: baseline } = await pg.query(
      "SELECT 1 FROM goose_db_version WHERE version_id = 20261019200000 AND is_applied"
    ).catch(() => ({ rows: [] }));
    if (baseline.length === 0) {
      throw new Error("Postgres baseline migration is not applied; run goose up first");
    }

    const existing = [];
    for (const table of TABLES) {
      const where = table === "box_systems" ? "WHERE user_id IS NOT NULL" : "";
      if (SEEDED_TABLES.has(table) && table !== "box_systems") continue;
      const { rows } = await pg.query(`SELECT COUNT(*)::int AS cnt FROM "${table}" ${where}`);
      if (rows[0].cnt > 0) {
        existing.push(`${table}=${rows[0].cnt}`);
      }
    }
    if (existing.length > 0) {
      console.log(`Postgres already has user data: ${existing.join(", ")}`);
      if (!FORCE) {
        throw new Error("Refusing to truncate Postgres tables with user data without FORCE=1");
      }
    }

    await pg.query("BEGIN");
    await pg.query(`TRUNCATE ${TABLES.map((t) => `"${t}"`).join(", ")} RESTART IDENTITY`);

    const summary = [];
    for (const table of TABLES) {
      const started = Date.now();
      const kinds = await pgColumnKinds(pg, table);
      const { columns, copied } = await copyTable(my, pg, table, kinds);
      const [[{ cnt: mysqlCount }]] = await my.query(`SELECT COUNT(*) AS cnt FROM \`${table}\``);
      const { rows } = await pg.query(`SELECT COUNT(*)::int AS cnt FROM "${table}"`);
      summary.push({ table, mysqlCount: Number(mysqlCount), copied, pgCount: rows[0].cnt, seconds: (Date.now() - started) / 1000 });
      await verifyContent(my, pg, table, columns, kinds);

      // setval is not transactional, so a dry run must not touch sequences
      if (!DRY_RUN) {
        await pg.query(
          `SELECT setval(pg_get_serial_sequence($1, 'id'), COALESCE((SELECT MAX(id) FROM "${table}"), 1), (SELECT MAX(id) IS NOT NULL FROM "${table}"))`,
          [table]
        );
      }
    }

    console.table(summary);
    const mismatches = summary.filter((s) => s.mysqlCount !== s.pgCount || s.copied !== s.pgCount);
    if (mismatches.length > 0) {
      throw new Error(`Row count mismatch: ${mismatches.map((m) => m.table).join(", ")}`);
    }
    console.log("Per-column content (non-null counts, sums, text lengths, timestamps, JSON element counts, booleans) matches MySQL");

    if (DRY_RUN) {
      await pg.query("ROLLBACK");
      console.log("DRY_RUN=1: copy verified and rolled back");
    } else {
      await pg.query("COMMIT");
      console.log("Data copy committed");
    }
  } catch (err) {
    await pg.query("ROLLBACK").catch(() => {});
    throw err;
  } finally {
    await my.end();
    await pg.end();
  }
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
