# FinTrack Coach — local developer commands
#
# Database URL points at the Docker Postgres published on host port 5433.
DATABASE_URL ?= postgres://fintrack:fintrack@localhost:5433/fintrack?sslmode=disable

# Run migrate from the backend module so go.mod pins the CLI version.
# -tags postgres includes the PostgreSQL database driver.
MIGRATE = go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate

.PHONY: migrate-up migrate-down migrate-version migrate-create

# Apply all pending migrations.
migrate-up:
	cd backend && $(MIGRATE) -path ./migrations -database "$(DATABASE_URL)" up

# Roll back the most recent migration.
migrate-down:
	cd backend && $(MIGRATE) -path ./migrations -database "$(DATABASE_URL)" down 1

# Print the current migration version.
migrate-version:
	cd backend && $(MIGRATE) -path ./migrations -database "$(DATABASE_URL)" version

# Create a new timestamped migration pair.
# Usage: make migrate-create name=create_transactions
migrate-create:
	@test -n "$(name)" || (echo "Usage: make migrate-create name=<migration_name>"; exit 1)
	cd backend && $(MIGRATE) create -ext sql -dir ./migrations -seq $(name)
