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

db-reset:
	@echo ">> dropping & recreating DB '$(PGDATABASE)' as $(PGADMIN_USER)@$(PGHOST)"
	@PGPASSWORD="$(PGADMIN_PASSWORD)" dropdb   -h "$(PGHOST)" -U "$(PGADMIN_USER)" "$(PGDATABASE)" || true
	@PGPASSWORD="$(PGADMIN_PASSWORD)" createdb -h "$(PGHOST)" -U "$(PGADMIN_USER)" -O "$(PGUSER)" "$(PGDATABASE)"
	@echo ">> fixando permissões do schema public para $(PGUSER)"
	@PGPASSWORD="$(PGADMIN_PASSWORD)" psql -h "$(PGHOST)" -U "$(PGADMIN_USER)" -d "$(PGDATABASE)" -v ON_ERROR_STOP=1 -c "ALTER SCHEMA public OWNER TO \"$(PGUSER)\";"
	@PGPASSWORD="$(PGADMIN_PASSWORD)" psql -h "$(PGHOST)" -U "$(PGADMIN_USER)" -d "$(PGDATABASE)" -v ON_ERROR_STOP=1 -c "GRANT ALL ON SCHEMA public TO \"$(PGUSER)\";"
	@PGPASSWORD="$(PGADMIN_PASSWORD)" psql -h "$(PGHOST)" -U "$(PGADMIN_USER)" -d "$(PGDATABASE)" -v ON_ERROR_STOP=1 -c "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO \"$(PGUSER)\";"
	@echo ">> enabling pgcrypto (as admin)"
	@PGPASSWORD="$(PGADMIN_PASSWORD)" psql -h "$(PGHOST)" -U "$(PGADMIN_USER)" -d "$(PGDATABASE)" -c "CREATE EXTENSION IF NOT EXISTS pgcrypto;"
	@$(MAKE) db-up


db-up:
	@migrate -path "$(MIGRATIONS_DIR)" -database "$(DATABASE_URL)" up

db-down:
	@migrate -path "$(MIGRATIONS_DIR)" -database "$(DATABASE_URL)" down 1

db-validate:
	@bash ./tools/validate_migrations.sh "$(DATABASE_URL)" "$(MIGRATIONS_DIR)" "$(TEST_DB)"

db-drift:
	@bash ./tools/drift_check.sh "$(DATABASE_URL)" "$(MIGRATIONS_DIR)"

# ===============================
# Database commands
# ===============================

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

