#!/usr/bin/env bash
set -euo pipefail

DATABASE_URL="${1:?DATABASE_URL required}"
MIGRATIONS_DIR="${2:?MIGRATIONS_DIR required}"

# --- parse DATABASE_URL ---
url="${DATABASE_URL#*://}"
creds="${url%@*}"
hostdb="${url#*@}"

DB_USER="${creds%%:*}"
DB_PASS="${creds#*:}"
DB_HOST="${hostdb%%[:/]*}"

rest="${hostdb#*:}"
if [[ "$rest" != "$hostdb" ]]; then
  DB_PORT="${rest%%/*}"
  DB_NAME="${rest#*/}"
else
  DB_PORT="5432"
  DB_NAME="${hostdb#*/}"
fi
DB_NAME="${DB_NAME%%\?*}"

BASE_URL_NO_DB="postgres://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}"

echo "== Drift check =="

TMP_DIR="$(mktemp -d)"
REAL_DUMP="$TMP_DIR/real.sql"
REBUILT_DUMP="$TMP_DIR/rebuilt.sql"

# Dump do banco real
PGPASSWORD="$DB_PASS" pg_dump -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" --schema-only > "$REAL_DUMP"

# Prepara DB temporário
TEST_DB="anima_drift_$$"
TEST_URL="${BASE_URL_NO_DB}/${TEST_DB}?sslmode=disable"

PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -U "$DB_USER" -d postgres -c "DROP DATABASE IF EXISTS ${TEST_DB};"
PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -U "$DB_USER" -d postgres -c "CREATE DATABASE ${TEST_DB};"

# Sobe schema via migrations e dumpa
migrate -path "$MIGRATIONS_DIR" -database "$TEST_URL" up
PGPASSWORD="$DB_PASS" pg_dump -h "$DB_HOST" -U "$DB_USER" -d "$TEST_DB" --schema-only > "$REBUILT_DUMP"

# Normaliza comentários/ruído
sed -i 's/--.*//g' "$REAL_DUMP"
sed -i 's/--.*//g' "$REBUILT_DUMP"

echo "-> diff:"
if diff -u "$REBUILT_DUMP" "$REAL_DUMP"; then
  echo "Sem drift detectado."
else
  echo "DRIFT detectado! Ajuste migrations para alinhar schema."
  exit 1
fi
