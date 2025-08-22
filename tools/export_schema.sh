#!/usr/bin/env bash
set -euo pipefail

# ===== Config padrão (pode sobrescrever por env) =====
PGHOST="${PGHOST:-127.0.0.1}"
PGPORT="${PGPORT:-5432}"
PGUSER="${PGUSER:-anima}"
PGPASSWORD="${PGPASSWORD:-anima}"
PGDATABASE="${PGDATABASE:-anima}"
SCHEMA="${SCHEMA:-public}"

OUT_DIR="${OUT_DIR:-docs}"
SQL_OUT="$OUT_DIR/schema.sql"
MD_OUT="$OUT_DIR/schema.md"

export PGPASSWORD

require_tools() {
  command -v psql >/dev/null || { echo "psql não encontrado"; exit 127; }
  command -v pg_dump >/dev/null || { echo "pg_dump não encontrado"; exit 127; }
}

init_out() {
  mkdir -p "$OUT_DIR"
  : > "$SQL_OUT"
  : > "$MD_OUT"
}

header_md() {
  local safe_user="$PGUSER"
  local safe_host="$PGHOST"
  local safe_db="$PGDATABASE"
  {
    echo "# Anima – Snapshot de Schema"
    echo ""
    echo "- Host: \`$safe_host:$PGPORT\`"
    echo "- Database: \`$safe_db\`"
    echo "- Schema: \`$SCHEMA\`"
    echo "- Data: \`$(date -Iseconds)\`"
    echo ""
    echo "> Gerado por tools/export_schema.sh"
    echo ""
  } >> "$MD_OUT"
}

dump_schema_sql() {
  echo ">> Gerando $SQL_OUT"
  # Exclui tabelas de controle comuns de ferramentas de migration
  local excludes=( "--exclude-table=${SCHEMA}.schema_migrations"
                   "--exclude-table=${SCHEMA}.goose_db_version"
                   "--exclude-table=${SCHEMA}.alembic_version"
                   "--exclude-table=${SCHEMA}.gorp_migrations" )
  pg_dump \
    --schema="${SCHEMA}" \
    --schema-only --no-owner --no-privileges --no-comments \
    "${excludes[@]}" \
    --host "$PGHOST" --port "$PGPORT" --username "$PGUSER" "$PGDATABASE" \
    > "$SQL_OUT"
}

# Executa uma query e imprime linhas no formato "a|b|c"
q() {
  psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -At -F '|' -c "$1"
}

list_tables() {
  q "SELECT table_name
     FROM information_schema.tables
     WHERE table_schema='${SCHEMA}' AND table_type='BASE TABLE'
     ORDER BY table_name;"
}

table_columns() {
  local tbl="$1"
  q "SELECT c.column_name,
            c.data_type,
            COALESCE(c.character_maximum_length::text,''),
            c.is_nullable,
            COALESCE(c.column_default,'')
     FROM information_schema.columns c
     WHERE c.table_schema='${SCHEMA}'
       AND c.table_name='${tbl}'
     ORDER BY c.ordinal_position;"
}

table_pk() {
  local tbl="$1"
  q "SELECT kcu.column_name
     FROM information_schema.table_constraints tc
     JOIN information_schema.key_column_usage kcu
       ON tc.constraint_name=kcu.constraint_name
      AND tc.table_schema=kcu.table_schema
     WHERE tc.table_schema='${SCHEMA}'
       AND tc.table_name='${tbl}'
       AND tc.constraint_type='PRIMARY KEY'
     ORDER BY kcu.ordinal_position;"
}

table_fks() {
  local tbl="$1"
  q "SELECT conname, pg_get_constraintdef(oid)
     FROM pg_constraint
     WHERE conrelid = format('${SCHEMA}.%I', '${tbl}')::regclass
       AND contype = 'f'
     ORDER BY conname;"
}

table_indexes() {
  local tbl="$1"
  q "SELECT indexname, indexdef
     FROM pg_indexes
     WHERE schemaname='${SCHEMA}' AND tablename='${tbl}'
     ORDER BY indexname;"
}

emit_table_md() {
  local tbl="$1"
  {
    echo "## ${tbl}"
    echo ""
    echo "### Colunas"
    echo ""
    echo "| coluna | tipo | len | null | default |"
    echo "|-------|------|-----|------|---------|"
  } >> "$MD_OUT"

  local row
  while IFS='|' read -r col dtype clen null def; do
    [[ -z "$col" ]] && continue
    [[ "$clen" == "" ]] || clen="($clen)"
    # Escapa barras verticais que possam aparecer em defaults/definições
    def="${def//|/\\|}"
    echo "| \`$col\` | \`$dtype\` | \`$clen\` | \`$null\` | \`$def\` |" >> "$MD_OUT"
  done < <(table_columns "$tbl")

  # PK
  {
    echo ""
    echo "### Primary Key"
    local pk="$(table_pk "$tbl" | paste -sd ', ' -)"
    [[ -z "$pk" ]] && pk="(nenhuma)"
    echo ""
    echo "- $pk"
  } >> "$MD_OUT"

  # FKs
  {
    echo ""
    echo "### Foreign Keys"
    local fk_any=0
    while IFS='|' read -r name def; do
      if [[ -n "$name" ]]; then
        fk_any=1
        def="${def//|/\\|}"
        echo "- **$name**: \`$def\`" >> "$MD_OUT"
      fi
    done < <(table_fks "$tbl")
    [[ $fk_any -eq 0 ]] && echo "- (nenhuma)" >> "$MD_OUT"
  } >> "$MD_OUT"

  # Indexes
  {
    echo ""
    echo "### Índices"
    local idx_any=0
    while IFS='|' read -r name def; do
      if [[ -n "$name" ]]; then
        idx_any=1
        def="${def//|/\\|}"
        echo "- **$name**: \`$def\`" >> "$MD_OUT"
      fi
    done < <(table_indexes "$tbl")
    [[ $idx_any -eq 0 ]] && echo "- (nenhum)" >> "$MD_OUT"
    echo ""
  } >> "$MD_OUT"
}

emit_extensions_md() {
  {
    echo "## Extensões"
    echo ""
  } >> "$MD_OUT"
  psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -At -F '|' \
    -c "SELECT extname, extversion FROM pg_extension ORDER BY extname;" \
    | while IFS='|' read -r name ver; do
        [[ -z "$name" ]] && continue
        echo "- \`$name\` (\`$ver\`)" >> "$MD_OUT"
      done
  echo "" >> "$MD_OUT"
}

main() {
  require_tools
  init_out
  header_md
  dump_schema_sql

  echo ">> Gerando $MD_OUT"
  emit_extensions_md

  # Lista tabelas e emite markdown para cada uma
  while IFS= read -r tbl; do
    [[ -z "$tbl" ]] && continue
    emit_table_md "$tbl"
  done < <(list_tables)

  echo "Concluído:"
  echo " - $SQL_OUT"
  echo " - $MD_OUT"
}

main "$@"
