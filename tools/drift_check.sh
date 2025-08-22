#!/usr/bin/env bash
set -euo pipefail

# ===============================
# Defaults
# ===============================
PGHOST="${PGHOST:-127.0.0.1}"
PGPORT="${PGPORT:-5432}"
PGUSER="${PGUSER:-anima}"
PGPASSWORD="${PGPASSWORD:-anima}"
PGDATABASE="${PGDATABASE:-anima}"

DB_USER="${DB_USER:-$PGUSER}"
DB_PASS="${DB_PASS:-$PGPASSWORD}"
DB_NAME="${DB_NAME:-$PGDATABASE}"

MIGRATIONS_DIR="${MIGRATIONS_DIR:-db/migrations}"

export PGPASSWORD

usage() { echo "Usage: $0 <validate|drift|reset>"; }

require_psql() {
  command -v psql >/dev/null || { echo "psql não encontrado"; exit 127; }
  command -v pg_dump >/dev/null || { echo "pg_dump não encontrado"; exit 127; }
}

conn_url() { local db="$1"; echo "postgres://${DB_USER}:${DB_PASS}@${PGHOST}:${PGPORT}/${db}?sslmode=disable"; }

psql_run() {
  local db="$1"; shift || true
  PGPASSWORD="$DB_PASS" psql "$(conn_url "$db")" -v ON_ERROR_STOP=1 "$@"
}

drop_db_if_exists() {
  local db="$1"
  psql -h "$PGHOST" -p "$PGPORT" -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 -c \
    "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${db}' AND pid <> pg_backend_pid();" >/dev/null || true
  psql -h "$PGHOST" -p "$PGPORT" -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 -c "DROP DATABASE IF EXISTS ${db};"
}

create_db() {
  local db="$1"
  psql -h "$PGHOST" -p "$PGPORT" -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 -c "CREATE DATABASE ${db} OWNER ${DB_USER};"
}

# Só aplica UP/SEED. Nunca *.down.sql
apply_migrations() {
  local db="$1"
  [[ -d "$MIGRATIONS_DIR" ]] || { echo "Diretório de migrations não encontrado: $MIGRATIONS_DIR"; exit 2; }

  mapfile -t files < <(
    find "$MIGRATIONS_DIR" -maxdepth 1 -type f \
      \( -name "*.up.sql" -o -name "*seed*.sql" -o -name "*_pg.sql" -o -name "*_postgres.sql" \) \
      | sort
  )
  [[ ${#files[@]} -gt 0 ]] || { echo "Nenhuma migration válida em $MIGRATIONS_DIR"; exit 3; }

  echo "Aplicando ${#files[@]} migrations em '$db'..."
  for f in "${files[@]}"; do
    echo ">> $f"
    psql_run "$db" -f "$f"
  done
}

cmd_validate() {
  local tmp_db="${DB_NAME}_migrate_check"
  echo "== Validate migrations =="
  drop_db_if_exists "$tmp_db"
  create_db "$tmp_db"
  apply_migrations "$tmp_db"
  echo "OK: migrations aplicaram com sucesso."
  drop_db_if_exists "$tmp_db"
}

cmd_drift() {
  local tmp_db="${DB_NAME}_driftcheck"
  local dump_tmp; dump_tmp="$(mktemp)"
  local dump_live; dump_live="$(mktemp)"

  echo "== Drift check =="
  drop_db_if_exists "$tmp_db"
  create_db "$tmp_db"
  apply_migrations "$tmp_db"

    echo "Gerando dumps de schema..."
  EXCLUDES=(
    "--exclude-table=schema_migrations"
    "--exclude-table=goose_db_version"
    "--exclude-table=alembic_version"
    "--exclude-table=gorp_migrations"
  )

  # Dump do DB temporário, removendo comentários e linhas vazias
  PGPASSWORD="$DB_PASS" pg_dump --schema-only --no-owner --no-privileges --no-comments \
    "${EXCLUDES[@]}" \
    --host "$PGHOST" --port "$PGPORT" --username "$DB_USER" "$tmp_db" \
    | sed -E '/^--/d;/^[[:space:]]*$/d' > "$dump_tmp"

  # Dump do DB live, removendo comentários e linhas vazias
  PGPASSWORD="$DB_PASS" pg_dump --schema-only --no-owner --no-privileges --no-comments \
    "${EXCLUDES[@]}" \
    --host "$PGHOST" --port "$PGPORT" --username "$DB_USER" "$DB_NAME" \
    | sed -E '/^--/d;/^[[:space:]]*$/d' > "$dump_live"

  if diff -u "$dump_live" "$dump_tmp" > /tmp/drift.diff; then
    echo "Sem drift detectado."
    rm -f "$dump_tmp" "$dump_live" /tmp/drift.diff
    drop_db_if_exists "$tmp_db"
    exit 0
  else
    echo "DRIFT detectado! Veja /tmp/drift.diff"
    echo "Preview (até 120 linhas):"
    head -n 120 /tmp/drift.diff || true
    rm -f "$dump_tmp" "$dump_live"
    drop_db_if_exists "$tmp_db"
    exit 1
  fi
}

cmd_reset() {
  echo "== Reset local DB '${DB_NAME}' =="
  drop_db_if_exists "$DB_NAME"
  create_db "$DB_NAME"
  apply_migrations "$DB_NAME"
  echo "OK: DB '${DB_NAME}' recriado e migrado."
}

main() {
  require_psql
  case "${1:-}" in
    validate) cmd_validate ;;
    drift)    cmd_drift ;;
    reset)    cmd_reset ;;
    *) usage; exit 64 ;;
  esac
}
main "$@"
