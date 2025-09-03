// Represents the status of a data consumer's application
export type ApplicationStatus = "pending" | "approved" | "denied";

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
export type ProviderType = "government" | "board" | "business";

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
export type ProviderSchemaStatus = "pending" | "approved" | "changes_required";

/**
 * Represents a data provider's submission, including their schema and field metadata
 */
export interface ProviderSchema {
  submissionId: string;
  providerId: string;
  status: ProviderSchemaStatus;
  apiEndpoint: string;
  schema: string;
  fieldMetadata: object;
}
