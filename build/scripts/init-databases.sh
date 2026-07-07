#!/bin/bash
set -euo pipefail

# Create one database per service. Idempotent, and skips the cluster's default
# POSTGRES_DB so it never collides with a service that shares that name.
for db in auth catalog reservation checkout ticketing notification analytics; do
  if [ "$db" = "${POSTGRES_DB:-}" ]; then
    echo "database already exists (POSTGRES_DB): $db — skipping"
    continue
  fi
  exists=$(psql -tAc "SELECT 1 FROM pg_database WHERE datname = '$db'" --username "$POSTGRES_USER")
  if [ "$exists" = "1" ]; then
    echo "database already exists: $db — skipping"
    continue
  fi
  echo "creating database: $db"
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" -c "CREATE DATABASE $db"
done
