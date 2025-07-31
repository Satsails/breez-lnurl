#!/bin/sh
# Exit immediately if a command exits with a non-zero status.
set -e

echo "Running database migrations..."
# The 'migrate' tool connects to the database using the DATABASE_URL env var.
# It checks the 'schema_migrations' table to see which migrations have been applied.
# The 'up' command is idempotent: it only applies migrations that have not been run yet.
./migrate -path ./persist/migrations -database "$DATABASE_URL" -verbose up

echo "Migrations complete."

echo "Starting application..."
./breez-lnurl