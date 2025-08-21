export $(shell sed 's/=.*//' .env)
MIGRATIONS_DIR=./db/migrations
.PHONY: db-up db-down db-reset db-validate db-drift ci-validate
db-up:      ; migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up
db-down:    ; migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1
db-reset:
@db_name=$$(echo "$(DATABASE_URL)" | sed -E 's#.*/([^?]+).*#\1#'); \
psql "$(DATABASE_URL)" -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='$$db_name' AND pid <> pg_backend_pid();" || true; \
psql "$(DATABASE_URL)" -d postgres -c "DROP DATABASE IF EXISTS $$db_name;" ; \
psql "$(DATABASE_URL)" -d postgres -c "CREATE DATABASE $$db_name;"
@$(MAKE) db-up
db-validate: ; bash ./tools/validate_migrations.sh "$(DATABASE_URL)" "$(MIGRATIONS_DIR)" "$(TEST_DB)"
db-drift:    ; bash ./tools/drift_check.sh "$(DATABASE_URL)" "$(MIGRATIONS_DIR)"
ci-validate: ; $(MAKE) db-validate db-drift
