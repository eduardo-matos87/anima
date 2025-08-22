# --- carrega .env se existir ---
ifneq (,$(wildcard .env))
include .env
export $(shell sed -n 's/^\([^#=[:space:]]\+\)=.*/\1/p' .env)
endif

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
PGHOST     ?= 127.0.0.1
PGPORT     ?= 5432
PGUSER     ?= anima
PGPASSWORD ?= anima
PGDATABASE ?= anima

export PGHOST PGPORT PGUSER PGPASSWORD PGDATABASE

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

