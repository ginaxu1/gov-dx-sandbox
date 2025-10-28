#!/bin/bash

# Database setup script for audit service using remote PostgreSQL
# This script runs the migration to create the audit_logs table and related structures

set -e  # Exit on any error

echo "🗄️  Setting up Audit Service Database (Remote PostgreSQL)"
echo "========================================================"

# Database configuration from environment variables
DB_HOST=${CHOREO_DB_AUDIT_HOSTNAME:-localhost}
DB_PORT=${CHOREO_DB_AUDIT_PORT:-5432}
DB_USER=${CHOREO_DB_AUDIT_USERNAME:-user}
DB_PASSWORD=${CHOREO_DB_AUDIT_PASSWORD:-password}
DB_NAME=${CHOREO_DB_AUDIT_DATABASENAME:-defaultdb}
DB_SSLMODE=${DB_SSLMODE:-require}

echo "Database: $DB_HOST:$DB_PORT/$DB_NAME"
echo "User: $DB_USER"
echo "SSL Mode: $DB_SSLMODE"

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo "❌ psql command not found. Please install PostgreSQL client tools."
    exit 1
fi

# Test database connection
echo "Testing database connection..."
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT version();" > /dev/null

if [ $? -eq 0 ]; then
    echo "✅ Database connection successful!"
else
    echo "❌ Database connection failed!"
    exit 1
fi

# Run the minimal setup
echo "Running minimal database setup..."
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f database/minimal-setup.sql

if [ $? -eq 0 ]; then
    echo "✅ Minimal database setup completed successfully!"
else
    echo "❌ Database setup failed!"
    exit 1
fi

# Verify the setup
echo "Verifying database setup..."
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
SELECT 
    'audit_logs' as table_name, 
    COUNT(*) as record_count 
FROM audit_logs
UNION ALL
SELECT 
    'consumer_applications' as table_name, 
    COUNT(*) as record_count 
FROM consumer_applications
UNION ALL
SELECT 
    'provider_schemas' as table_name, 
    COUNT(*) as record_count 
FROM provider_schemas;
"

echo "✅ Database setup completed successfully!"
echo "The audit service should now be able to query audit logs."
