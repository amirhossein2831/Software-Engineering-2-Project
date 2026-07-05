#!/bin/bash
set -euo pipefail

for db in auth catalog reservation checkout ticketing notification analytics; do
  echo "creating database: $db"
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE $db;
EOSQL
done
