-- Database initialization for integration tests
-- This creates the necessary tables for all services

-- API Server Tables
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS entities (
    entity_id VARCHAR(255) PRIMARY KEY,
    entity_name VARCHAR(255) NOT NULL,
    contact_email VARCHAR(255) NOT NULL,
    phone_number VARCHAR(50),
    entity_type VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS consumers (
    consumer_id VARCHAR(255) PRIMARY KEY,
    entity_id VARCHAR(255) NOT NULL REFERENCES entities(entity_id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS consumer_apps (
    submission_id VARCHAR(255) PRIMARY KEY,
    consumer_id VARCHAR(255) NOT NULL REFERENCES consumers(consumer_id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    required_fields JSONB,
    credentials JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS provider_submissions (
    submission_id VARCHAR(255) PRIMARY KEY,
    provider_name VARCHAR(255) NOT NULL,
    contact_email VARCHAR(255) NOT NULL,
    phone_number VARCHAR(50) NOT NULL,
    provider_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS provider_profiles (
    provider_id VARCHAR(255) PRIMARY KEY,
    entity_id VARCHAR(255) NOT NULL,
    provider_name VARCHAR(255) NOT NULL,
    contact_email VARCHAR(255) NOT NULL,
    phone_number VARCHAR(50) NOT NULL,
    provider_type VARCHAR(100) NOT NULL,
    approved_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS provider_schemas (
    submission_id VARCHAR(255) PRIMARY KEY,
    provider_id VARCHAR(255) NOT NULL REFERENCES provider_profiles(provider_id) ON DELETE CASCADE,
    schema_id VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    schema_input JSONB,
    sdl TEXT,
    schema_endpoint VARCHAR(500),
    field_configurations JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS consumer_grants (
    consumer_id VARCHAR(255) PRIMARY KEY,
    approved_fields JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS provider_metadata (
    field_name VARCHAR(255) PRIMARY KEY,
    owner VARCHAR(255) NOT NULL,
    provider VARCHAR(255) NOT NULL,
    consent_required BOOLEAN NOT NULL DEFAULT false,
    access_control_type VARCHAR(100) NOT NULL DEFAULT 'public',
    allow_list JSONB,
    description TEXT,
    expiry_time VARCHAR(50),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_consumer_apps_consumer_id ON consumer_apps(consumer_id);
CREATE INDEX IF NOT EXISTS idx_provider_schemas_provider_id ON provider_schemas(provider_id);
CREATE INDEX IF NOT EXISTS idx_consumers_entity_id ON consumers(entity_id);

-- Consent Engine Tables
CREATE TABLE IF NOT EXISTS consents (
    consent_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    consumer_id VARCHAR(255) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    fields TEXT[] NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    otp_code VARCHAR(10),
    otp_expiry TIMESTAMP WITH TIME ZONE,
    owner_phone VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_consents_consumer_id ON consents(consumer_id);
CREATE INDEX IF NOT EXISTS idx_consents_status ON consents(status);

-- Audit Service Tables
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) NOT NULL,
    requested_data JSONB,
    application_id VARCHAR(255),
    schema_id VARCHAR(255),
    consumer_id VARCHAR(255),
    provider_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_logs_consumer_id ON audit_logs(consumer_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_provider_id ON audit_logs(provider_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_status ON audit_logs(status);

-- Policy Decision Point Tables
CREATE TABLE IF NOT EXISTS policy_metadata (
    field_name VARCHAR(255) PRIMARY KEY,
    owner VARCHAR(255) NOT NULL,
    provider VARCHAR(255) NOT NULL,
    consent_required BOOLEAN NOT NULL DEFAULT false,
    access_control_type VARCHAR(100) NOT NULL DEFAULT 'public',
    allow_list JSONB,
    description TEXT,
    expiry_time VARCHAR(50),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_policy_metadata_provider ON policy_metadata(provider);
CREATE INDEX IF NOT EXISTS idx_policy_metadata_owner ON policy_metadata(owner);

-- Orchestration Engine Tables
CREATE TABLE IF NOT EXISTS unified_schemas (
    id VARCHAR(36) PRIMARY KEY,
    version VARCHAR(50) UNIQUE NOT NULL,
    sdl TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'inactive',
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(100),
    checksum VARCHAR(64) NOT NULL,
    is_active BOOLEAN DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_unified_schemas_version ON unified_schemas(version);
CREATE INDEX IF NOT EXISTS idx_unified_schemas_status ON unified_schemas(status);

CREATE TABLE IF NOT EXISTS schema_versions (
    id SERIAL PRIMARY KEY,
    from_version VARCHAR(50),
    to_version VARCHAR(50) NOT NULL,
    change_type VARCHAR(20) NOT NULL,
    changes JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL
);

