-- Seed data for integration tests
-- This file populates the database with known test data

-- Insert test entities
INSERT INTO entities (entity_id, entity_name, contact_email, phone_number, entity_type)
VALUES 
    ('entity-1', 'Test Entity', 'test@example.com', '+1234567890', 'government'),
    ('entity-2', 'Registry of Persons', 'rgd@example.com', '+1111111111', 'government'),
    ('entity-3', 'Department of Motor Transport', 'dmt@example.com', '+2222222222', 'government')
ON CONFLICT (entity_id) DO NOTHING;

-- Insert test consumers
INSERT INTO consumers (consumer_id, entity_id)
VALUES 
    ('passport-app', 'entity-1'),
    ('test-consumer', 'entity-1'),
    ('unauthorized-app', 'entity-2')
ON CONFLICT (consumer_id) DO NOTHING;

-- Insert test provider profiles
INSERT INTO provider_profiles (provider_id, entity_id, provider_name, contact_email, phone_number, provider_type)
VALUES 
    ('provider-drp', 'entity-3', 'Registry of Persons (DRP)', 'drp@example.com', '+3333333333', 'data-provider'),
    ('provider-rgd', 'entity-2', 'Registry Department', 'rgd@example.com', '+4444444444', 'data-provider')
ON CONFLICT (provider_id) DO NOTHING;

-- Insert test provider schemas
INSERT INTO provider_schemas (submission_id, provider_id, schema_id, status, sdl, schema_endpoint, field_configurations)
VALUES 
    ('schema-1', 'provider-drp', 'schema-person', 'approved', 
     'type Person { fullName: String! birthDate: String permanentAddress: String photo: String nic: String }',
     'http://provider-drp/graphql',
     '{}'::jsonb)
ON CONFLICT (submission_id) DO NOTHING;

-- Insert test provider metadata
INSERT INTO provider_metadata (field_name, owner, provider, consent_required, access_control_type, allow_list, description)
VALUES 
    -- Public field
    ('person.fullName', 'citizen', 'drp', false, 'public', '[]'::jsonb, 'Full name of the person'),
    
    -- Restricted field with allow_list
    ('person.birthDate', 'rgd', 'drp', false, 'restricted', 
     '[{"consumerId":"passport-app","expires_at":1757560679,"grant_duration":"30d"}]'::jsonb, 
     'Birth date of the person'),
    
    -- Restricted field requiring consent
    ('person.permanentAddress', 'rgd', 'drp', true, 'restricted',
     '[{"consumerId":"passport-app","expires_at":1757560679,"grant_duration":"30d"}]'::jsonb,
     'Permanent address'),
    
    -- Restricted field not in allow_list
    ('person.nic', 'citizen', 'drp', true, 'restricted', '[]'::jsonb,
     'National Identity Card number')
ON CONFLICT (field_name) DO NOTHING;

-- Insert test consumer grants
INSERT INTO consumer_grants (consumer_id, approved_fields)
VALUES 
    ('passport-app', 
     '["person.fullName","person.birthDate","person.permanentAddress"]'::jsonb)
ON CONFLICT (consumer_id) DO NOTHING;

-- Insert test consents (for consent engine)
INSERT INTO consents (consent_id, consumer_id, provider_id, fields, status, owner_phone)
VALUES 
    (uuid_generate_v4(), 'test-consumer', 'provider-rgd', 
     ARRAY['person.permanentAddress'], 'approved', '+94777123456'),
    (uuid_generate_v4(), 'test-consumer', 'provider-drp',
     ARRAY['person.photo'], 'pending', '+94777123456')
ON CONFLICT (consent_id) DO NOTHING;

-- Insert test unified schemas (for orchestration engine)
INSERT INTO unified_schemas (id, version, sdl, status, description, created_by, checksum, is_active)
VALUES 
    ('schema-1', '1.0.0',
     'type Query { personInfo(nic: String!): Person } type Person { fullName: String birthDate: String permanentAddress: String photo: String nic: String }',
     'active', 'Test schema for integration tests', 'test-user', 'abc123', true)
ON CONFLICT (version) DO NOTHING;

