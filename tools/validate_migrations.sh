#!/usr/bin/env bash
set -euo pipefail
DATABASE_URL="${1:?}"; MIGRATIONS_DIR="${2:?}"; TEST_DB="${3:-anima_migrate_check}"
BASE_URL_NO_DB="$(echo "$DATABASE_URL" | sed -E 's#(.*//[^/]+)/(.*)#\1#')"
psql "$DATABASE_URL" -d postgres -c "DROP DATABASE IF EXISTS $TEST_DB;" >/dev/null
psql "$DATABASE_URL" -d postgres -c "CREATE DATABASE $TEST_DB;" >/dev/null
TEST_URL="${BASE_URL_NO_DB}/${TEST_DB}?sslmode=disable"
migrate -path "$MIGRATIONS_DIR" -database "$TEST_URL" up
psql "$TEST_URL" -v ON_ERROR_STOP=1 <<'SQL'
SELECT to_regclass('public.users')      IS NOT NULL AS has_users;
SELECT to_regclass('public.exercises')  IS NOT NULL AS has_exercises;
SELECT to_regclass('public.workouts')   IS NOT NULL AS has_workouts;
SQL
migrate -path "$MIGRATIONS_DIR" -database "$TEST_URL" down -all
migrate -path "$MIGRATIONS_DIR" -database "$TEST_URL" up
echo "OK: validação concluída."
