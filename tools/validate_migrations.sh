#!/usr/bin/env bash
set -euo pipefail

DATABASE_URL="${1:?DATABASE_URL required}"
MIGRATIONS_DIR="${2:?MIGRATIONS_DIR required}"
TEST_DB="${3:-anima_migrate_check}"

# --- parse DATABASE_URL: postgres://user:pass@host:port/dbname?qs ---
url="${DATABASE_URL#*://}"            # user:pass@host:port/db?params
creds="${url%@*}"                     # user:pass
hostdb="${url#*@}"                    # host:port/db?params

DB_USER="${creds%%:*}"
DB_PASS="${creds#*:}"
DB_HOST="${hostdb%%[:/]*}"

rest="${hostdb#*:}"
if [[ "$rest" != "$hostdb" ]]; then   # tem porta
  DB_PORT="${rest%%/*}"
  DB_NAME="${rest#*/}"
else
  DB_PORT="5432"
  DB_NAME="${hostdb#*/}"
fi
DB_NAME="${DB_NAME%%\?*}"

BASE_URL_NO_DB="postgres://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}"
TEST_URL="${BASE_URL_NO_DB}/${TEST_DB}?sslmode=disable"

echo "== Validate migrations =="
echo "-> Recriando DB de teste: $TEST_DB"
PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -U "$DB_USER" -d postgres -c "DROP DATABASE IF EXISTS ${TEST_DB};"
: # no-op
PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -U "$DB_USER" -d postgres -c "CREATE DATABASE ${TEST_DB};"

echo "-> migrate UP no DB de teste"
migrate -path "$MIGRATIONS_DIR" -database "$TEST_URL" up

echo "-> sanity checks"
PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -U "$DB_USER" -d "$TEST_DB" -v ON_ERROR_STOP=1 <<'SQL'
SELECT to_regclass('public.users')      IS NOT NULL AS has_users;
SELECT to_regclass('public.exercises')  IS NOT NULL AS has_exercises;
SELECT to_regclass('public.workouts')   IS NOT NULL AS has_workouts;
SQL

echo "-> migrate DOWN até zerar (ignora se não houver .down)"
if migrate -path "$MIGRATIONS_DIR" -database "$TEST_URL" down -all; then
  echo "down aplicado"
else
  echo "no down migrations, skipping"
fi

echo "-> subir novamente"
migrate -path "$MIGRATIONS_DIR" -database "$TEST_URL" up

echo "OK: validação concluída."
