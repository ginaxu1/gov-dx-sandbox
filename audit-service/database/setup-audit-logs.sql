-- Simplified Audit Logs Database Setup
-- This file creates the audit_logs table for PostgreSQL databases
-- Run this script to set up the audit service database schema

-- Create the audit_logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    consumer_id TEXT NOT NULL,
    provider_id TEXT,
    requested_data JSONB,
    response_data JSONB,
    transaction_status TEXT NOT NULL,
    citizen_hash TEXT,
    user_agent TEXT,
    ip_address INET,
    application_id VARCHAR(255),
    schema_id VARCHAR(255),
    status VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_audit_logs_consumer_id ON audit_logs(consumer_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_application_id ON audit_logs(application_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_schema_id ON audit_logs(schema_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- Create a view for easy querying with provider and consumer information
CREATE OR REPLACE VIEW audit_logs_with_provider_consumer AS
SELECT 
    al.id,
    al.event_id,
    al.timestamp,
    al.consumer_id,
    al.provider_id,
    al.requested_data,
    al.response_data,
    al.transaction_status,
    al.citizen_hash,
    al.user_agent,
    al.ip_address,
    al.application_id,
    al.schema_id,
    al.status,
    al.created_at,
    al.updated_at,
    -- Add provider and consumer names if needed (can be joined with other tables)
    al.consumer_id as consumer_name,
    al.provider_id as provider_name
FROM audit_logs al;

-- Grant permissions (adjust as needed for your environment)
-- GRANT SELECT, INSERT, UPDATE, DELETE ON audit_logs TO audit_service_user;
-- GRANT SELECT ON audit_logs_with_provider_consumer TO audit_service_user;

-- Optional: Create additional tables for consumer and provider metadata
CREATE TABLE IF NOT EXISTS consumer_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS provider_schemas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schema_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for the additional tables
CREATE INDEX IF NOT EXISTS idx_consumer_applications_application_id ON consumer_applications(application_id);
CREATE INDEX IF NOT EXISTS idx_provider_schemas_schema_id ON provider_schemas(schema_id);

-- Insert some sample data (optional)
INSERT INTO consumer_applications (application_id, name, description) 
VALUES ('app-123', 'Test Application', 'Sample application for testing')
ON CONFLICT (application_id) DO NOTHING;

INSERT INTO provider_schemas (schema_id, name, description) 
VALUES ('unknown-schema', 'Unknown Schema', 'Default schema for unknown providers')
ON CONFLICT (schema_id) DO NOTHING;

-- Verify the setup
SELECT 'Audit logs table created successfully' as status;
SELECT COUNT(*) as audit_logs_count FROM audit_logs;
SELECT COUNT(*) as consumer_applications_count FROM consumer_applications;
SELECT COUNT(*) as provider_schemas_count FROM provider_schemas;
