-- Performance optimization indexes for Policy Decision Point

-- Policy Metadata table indexes (additional to existing unique index)
CREATE INDEX IF NOT EXISTS idx_policy_metadata_schema_id ON policy_metadata(schema_id);
CREATE INDEX IF NOT EXISTS idx_policy_metadata_field_name ON policy_metadata(field_name);
CREATE INDEX IF NOT EXISTS idx_policy_metadata_source ON policy_metadata(source);
CREATE INDEX IF NOT EXISTS idx_policy_metadata_access_control_type ON policy_metadata(access_control_type);
CREATE INDEX IF NOT EXISTS idx_policy_metadata_is_owner ON policy_metadata(is_owner);
CREATE INDEX IF NOT EXISTS idx_policy_metadata_owner ON policy_metadata(owner);
CREATE INDEX IF NOT EXISTS idx_policy_metadata_created_at ON policy_metadata(created_at);
CREATE INDEX IF NOT EXISTS idx_policy_metadata_updated_at ON policy_metadata(updated_at);

-- Composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_policy_metadata_schema_field ON policy_metadata(schema_id, field_name);
CREATE INDEX IF NOT EXISTS idx_policy_metadata_schema_created ON policy_metadata(schema_id, created_at);
CREATE INDEX IF NOT EXISTS idx_policy_metadata_owner_created ON policy_metadata(owner, created_at);

-- GIN index for JSONB allow_list column for efficient JSON queries
CREATE INDEX IF NOT EXISTS idx_policy_metadata_allow_list_gin ON policy_metadata USING GIN (allow_list);
