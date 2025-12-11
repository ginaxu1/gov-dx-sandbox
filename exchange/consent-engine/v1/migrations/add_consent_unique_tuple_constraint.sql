-- Migration: Create consent_records table with composite unique constraint
-- Date: 2025-12-11
-- Description: Creates the consent_records table with consent_id as primary key
--              and a composite unique constraint on (owner_id, owner_email, app_id, created_at)

-- Create the consent_records table if it doesn't exist
CREATE TABLE IF NOT EXISTS consent_records (
    consent_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id VARCHAR(255) NOT NULL,
    owner_email VARCHAR(255) NOT NULL,
    app_id VARCHAR(255) NOT NULL,
    app_name VARCHAR(255),
    status VARCHAR(50) NOT NULL,
    type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    pending_expires_at TIMESTAMP WITH TIME ZONE,
    grant_expires_at TIMESTAMP WITH TIME ZONE,
    grant_duration VARCHAR(50) NOT NULL,
    fields JSONB NOT NULL,
    session_id VARCHAR(255),
    consent_portal_url TEXT NOT NULL,
    updated_by VARCHAR(255),
    
    -- Composite unique constraint
    CONSTRAINT idx_consent_unique_tuple UNIQUE (owner_id, owner_email, app_id, created_at)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_consent_records_owner_id ON consent_records(owner_id);
CREATE INDEX IF NOT EXISTS idx_consent_records_owner_email ON consent_records(owner_email);
CREATE INDEX IF NOT EXISTS idx_consent_records_app_id ON consent_records(app_id);
CREATE INDEX IF NOT EXISTS idx_consent_records_status ON consent_records(status);
CREATE INDEX IF NOT EXISTS idx_consent_records_created_at ON consent_records(created_at);
CREATE INDEX IF NOT EXISTS idx_consent_records_pending_expires_at ON consent_records(pending_expires_at);
CREATE INDEX IF NOT EXISTS idx_consent_records_grant_expires_at ON consent_records(grant_expires_at);

-- Composite index for owner and app lookups
CREATE INDEX IF NOT EXISTS idx_consent_records_owner_app ON consent_records(owner_id, app_id);

-- Add comment to table
COMMENT ON TABLE consent_records IS 'Stores consent records with unique constraint on (owner_id, owner_email, app_id, created_at)';

-- To rollback this migration (DROP TABLE):
-- DROP TABLE IF EXISTS consent_records CASCADE;
