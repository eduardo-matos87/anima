#!/usr/bin/env bash
set -euo pipefail
DATABASE_URL="${1:?}"; MIGRATIONS_DIR="${2:?}"
TMP_DIR="$(mktemp -d)"; REAL_DUMP="$TMP_DIR/real.sql"; REBUILT_DUMP="$TMP_DIR/rebuilt.sql"
pg_dump --schema-only "$DATABASE_URL" > "$REAL_DUMP"
TEST_DB="anima_drift_$$"; BASE_URL_NO_DB="$(echo "$DATABASE_URL" | sed -E 's#(.*//[^/]+)/(.*)#\1#')"
TEST_URL="${BASE_URL_NO_DB}/${TEST_DB}?sslmode=disable"
psql "$DATABASE_URL" -d postgres -c "DROP DATABASE IF EXISTS $TEST_DB;" >/dev/null
psql "$DATABASE_URL" -d postgres -c "CREATE DATABASE $TEST_DB;" >/dev/null
migrate -path "$MIGRATIONS_DIR" -database "$TEST_URL" up
pg_dump --schema-only "$TEST_URL" > "$REBUILT_DUMP"
sed -i 's/--.*//g' "$REAL_DUMP"; sed -i 's/--.*//g' "$REBUILT_DUMP"
if diff -u "$REBUILT_DUMP" "$REAL_DUMP"; then
  echo "Sem drift."
else
  echo "DRIFT detectado! Ajuste migrations."; exit 1
fi
