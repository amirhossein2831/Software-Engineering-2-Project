#!/bin/bash
# Creates one database per microservice on first cluster init
# (database-per-service). Runs automatically via docker-entrypoint-initdb.d.
set -euo pipefail

for db in iam catalog reservation checkout ticketing notification analytics; do
  echo "creating database: $db"
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE $db;
EOSQL
done
