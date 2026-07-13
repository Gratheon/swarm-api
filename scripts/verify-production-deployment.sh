#!/bin/sh
set -eu

EXPECTED_WORKING_DIR="${EXPECTED_WORKING_DIR:-/www/swarm-api}"
EXPECTED_COMMIT="${EXPECTED_COMMIT:-$(git -C "$EXPECTED_WORKING_DIR" rev-parse --short HEAD)}"
CONTAINER_NAME="${CONTAINER_NAME:-gratheon_swarm-api_1}"
SCHEMA_REGISTRY_URL="${SCHEMA_REGISTRY_URL:-http://127.0.0.1:3000/schema/latest}"

VERIFY_ATTEMPTS="${VERIFY_ATTEMPTS:-30}"
VERIFY_INTERVAL_SECONDS="${VERIFY_INTERVAL_SECONDS:-2}"

actual_working_dir=$(docker inspect "$CONTAINER_NAME" --format '{{ index .Config.Labels "com.docker.compose.project.working_dir" }}')
if [ "$actual_working_dir" != "$EXPECTED_WORKING_DIR" ]; then
  echo "Expected $CONTAINER_NAME to be deployed from $EXPECTED_WORKING_DIR, got $actual_working_dir" >&2
  exit 1
fi

for attempt in $(seq 1 "$VERIFY_ATTEMPTS"); do
  if python3 - "$EXPECTED_COMMIT" "$SCHEMA_REGISTRY_URL" <<'PY'
import json
import re
import sys
import urllib.request

expected_commit = sys.argv[1]
registry_url = sys.argv[2]
with urllib.request.urlopen(registry_url, timeout=10) as response:
    payload = json.load(response)
schema = next((entry for entry in payload.get("data", []) if entry.get("name") == "swarm-api"), None)
if schema is None:
    raise SystemExit("swarm-api schema is missing from registry")

actual_version = str(schema.get("version", "")).strip()
if actual_version != expected_commit:
    raise SystemExit(f"expected swarm-api schema version {expected_commit}, got {actual_version}")

required_fields = {
    "Query": ("boxSystems",),
    "Apiary": ("type",),
    "Hive": ("hiveType", "boxSystemId", "hiveNumber", "status", "lastInspection", "isNew", "families"),
    "Family": ("name", "age", "lastTreatment"),
    "Box": ("holeCount", "roofStyle"),
}
type_defs = schema.get("type_defs", "")
missing = []
for type_name, field_names in required_fields.items():
    match = re.search(rf"\btype\s+{re.escape(type_name)}\b[^{{]*{{(?P<body>.*?)}}", type_defs, re.DOTALL)
    if not match:
        missing.append(type_name)
        continue

    body = match.group("body")
    for field_name in field_names:
        if not re.search(rf"(?m)^\s*{re.escape(field_name)}\s*(?:\(|:)", body):
            missing.append(f"{type_name}.{field_name}")

if missing:
    raise SystemExit("swarm-api registry schema is missing contract entries: " + ", ".join(missing))

print(f"Verified swarm-api deployment {actual_version} from the expected working directory")
PY
  then
    exit 0
  fi

  if [ "$attempt" -lt "$VERIFY_ATTEMPTS" ]; then
    echo "Waiting for swarm-api schema registration (attempt $attempt/$VERIFY_ATTEMPTS)..." >&2
    sleep "$VERIFY_INTERVAL_SECONDS"
  fi
done

exit 1
