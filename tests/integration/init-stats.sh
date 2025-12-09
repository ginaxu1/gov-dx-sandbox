#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	CREATE DATABASE policy_db;
	CREATE DATABASE consent_db;
	CREATE DATABASE audit_db;
	CREATE DATABASE portal_db;
EOSQL
