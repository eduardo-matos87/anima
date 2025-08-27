# --- carrega .env se existir ---
ifneq (,$(wildcard .env))
include .env
export $(shell sed -n 's/^\([^#=[:space:]]\+\)=.*/\1/p' .env)
endif

<<<<<<< HEAD
# defaults
DATABASE_URL ?= postgres://anima:anima@localhost:5432/anima?sslmode=disable
PGHOST ?= localhost
PGUSER ?= anima
PGPASSWORD ?= anima
PGDATABASE ?= anima

MIGRATIONS_DIR=./db/migrations

.PHONY: run db-reset db-up db-down db-validate db-drift

run:
	go run ./...

.PHONY: db-validate db-drift db-reset db-create db-drop

# Variáveis padrão (podem ser sobrescritas no ambiente)
=======
SHELL := /bin/bash

# defaults
PORT          ?= 8081
DATABASE_URL  ?= postgres://anima:anima@localhost:5432/anima?sslmode=disable

>>>>>>> 948aba3 (Profiles + Metrics + Overload Admin + Infra (#1))
PGHOST     ?= 127.0.0.1
PGPORT     ?= 5432
PGUSER     ?= anima
PGPASSWORD ?= anima
PGDATABASE ?= anima

export PGHOST PGPORT PGUSER PGPASSWORD PGDATABASE

<<<<<<< HEAD
=======
MIGRATIONS_DIR = ./db/migrations

.PHONY: run db-reset db-up db-down db-validate db-drift db-create db-drop db-schema \
        db-refresh-overload psql

# ===== App =====
run:
	@PORT?=8081
	@DATABASE_URL?=postgres://anima:anima@localhost:5432/anima?sslmode=disable
	@echo "== Run API =="
	PORT=$$PORT DATABASE_URL=$$DATABASE_URL go run .


# ===== Banco / migrações utilitárias =====

>>>>>>> 948aba3 (Profiles + Metrics + Overload Admin + Infra (#1))
# Valida migrations criando DB de teste
db-validate:
	@echo "== Validate migrations =="
	@./tools/drift_check.sh validate

# Verifica drift entre schema esperado x atual
db-drift:
	@echo "== Drift check =="
	@./tools/drift_check.sh drift

# Reseta banco local (drop + create + migrate + seed)
db-reset:
	@echo "== Reset local DB =="
	@./tools/drift_check.sh reset

# Cria banco principal (se não existir)
db-create:
	@echo "== Create database =="
	@psql -h $(PGHOST) -p $(PGPORT) -U $(PGUSER) -c "CREATE DATABASE $(PGDATABASE);" || true

# Dropa banco principal
db-drop:
	@echo "== Drop database =="
	@psql -h $(PGHOST) -p $(PGPORT) -U $(PGUSER) -c "DROP DATABASE IF EXISTS $(PGDATABASE);" || true

<<<<<<< HEAD
.PHONY: db-schema
=======
# Exporta schema p/ docs/
>>>>>>> 948aba3 (Profiles + Metrics + Overload Admin + Infra (#1))
db-schema:
	@echo "== Exportando schema para docs/ =="
	@./tools/export_schema.sh

<<<<<<< HEAD

=======
# Abre psql rápido com a conexão atual
psql:
	@if [ -n "$$DATABASE_URL" ]; then \
		psql "$$DATABASE_URL"; \
	else \
		psql -h $(PGHOST) -p $(PGPORT) -U $(PGUSER) -d $(PGDATABASE); \
	fi

# ===== Overload: refresh da MV =====
# Primeiro tenta CONCURRENTLY; se for o primeiro populate (NO DATA), faz refresh normal.
db-refresh-overload:
	@echo "== Refresh MV workout_overload_stats12_mv =="
	@bash -c '\
		if [ -n "$$DATABASE_URL" ]; then \
			psql "$$DATABASE_URL" -c "REFRESH MATERIALIZED VIEW CONCURRENTLY workout_overload_stats12_mv;" || \
			psql "$$DATABASE_URL" -c "REFRESH MATERIALIZED VIEW workout_overload_stats12_mv;"; \
		else \
			psql -h $(PGHOST) -p $(PGPORT) -U $(PGUSER) -d $(PGDATABASE) -c "REFRESH MATERIALIZED VIEW CONCURRENTLY workout_overload_stats12_mv;" || \
			psql -h $(PGHOST) -p $(PGPORT) -U $(PGUSER) -d $(PGDATABASE) -c "REFRESH MATERIALIZED VIEW workout_overload_stats12_mv;"; \
		fi'

.PHONY: db-seed
db-seed:
	@echo "== Applying seeds in db/seeds (lexicographic order) =="
	@if [ -n "$$DATABASE_URL" ]; then \
		for f in $$(ls -1 db/seeds/*.sql 2>/dev/null | sort); do \
			echo ">> $$f"; psql "$$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$$f"; \
		done; \
	else \
		for f in $$(ls -1 db/seeds/*.sql 2>/dev/null | sort); do \
			echo ">> $$f"; psql -h $(PGHOST) -p $(PGPORT) -U $(PGUSER) -d $(PGDATABASE) -v ON_ERROR_STOP=1 -f "$$f"; \
		done; \
	fi
>>>>>>> 948aba3 (Profiles + Metrics + Overload Admin + Infra (#1))
