// Represents the status of a data consumer's application
export type ApplicationStatus = 'pending' | 'approved' | 'denied';

/**
 * Represents a data consumer's application to access specific data fields
 */
export interface Application {
  appId: string;
  status: ApplicationStatus;
  requiredFields: object;
  credentials?: {
    apiKey: string;
    apiSecret: string;
  };
}

// Represents the type of a data provider
export type ProviderType = 'government' | 'board' | 'business';

/**
 * Represents the core profile of a Data Provider organization
 */
export interface ProviderProfile {
    providerId: string;
    providerName: string;
    contactEmail: string;
    phoneNumber: string;
    providerType: ProviderType;
    createdAt: string;
}

// Represents the status of a data provider's schema submission
export type ProviderSchemaStatus = 'pending' | 'approved' | 'changes_required';

/**
 * @summary Defines the metadata for a single field in a provider's schema
 */
export interface FieldConfiguration {
    source: 'authoritative' | 'fallback' | 'other';
    isOwner: boolean;
    description: string;
}

/**
 * @summary Represents the nested structure of field configurations, grouped by GraphQL Type
 * @example
 * {
 * "PersonData": {
 * "nic": { source: 'fallback', isOwner: false, description: '...' },
 * "fullName": { source: 'authoritative', isOwner: true, description: '...' }
 * },
 * "Query": { ... }
 * }
 */
export type FieldConfigurations = Record<string, Record<string, FieldConfiguration>>;


/**
 * Represents a data provider's complete schema submission
 */
export interface ProviderSchema {
    submissionId: string;
    providerId: string; // Links to ProviderProfile
    status: ProviderSchemaStatus;
    // The original schema source (e.g., SDL, endpoint) - can be optional
    schemaInput?: {
        type: 'endpoint' | 'json' | 'sdl';
        value: string;
    };
    // The detailed, nested configuration for each field
    fieldConfigurations: FieldConfigurations;
}

