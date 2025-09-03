import express, { Request, Response } from 'express';
import crypto from 'crypto';

// TYPE DEFINITIONS (CONSUMER)

/** Represents the status of a data consumer's application. */
type ApplicationStatus = 'pending' | 'approved' | 'denied';

/**
 * Represents a data consumer's application to access specific data fields
 */
interface Application {
  /** A unique identifier for the application (e.g., "passport-app-123"). */
  appId: string;
  /** The current review status of the application. */
  status: ApplicationStatus;
  /** The data fields the application is requesting access to. */
  requiredFields: object;
  /** Secure credentials generated for the application upon approval. */
  credentials?: {
    apiKey: string;
    apiSecret: string;
  };
}

// TYPE DEFINITIONS (PROVIDER)

/** Represents the type of a data provider. */
type ProviderType = 'government' | 'board' | 'business';

/**
 * Represents the core profile of a Data Provider organization
 */
interface ProviderProfile {
    /** A system-generated unique identifier for the data provider. */
    providerId: string;
    /** The official name of the provider organization. */
    providerName: string;
    /** The primary contact email for the provider. */
    contactEmail: string;
    /** The contact phone number for the provider. */
    phoneNumber: string;
    /** The type of the provider organization. */
    providerType: ProviderType;
    /** The ISO timestamp when the provider was registered. */
    createdAt: string;
}

/** Represents the status of a data provider's schema submission. */
type ProviderSchemaStatus = 'pending' | 'approved' | 'changes_required';

/**
 * Represents a data provider's submission, including their schema and field metadata
 */
interface ProviderSchema {
    /** A unique identifier for this specific submission. */
    submissionId: string;
    /** A unique identifier for the data provider (links to ProviderProfile). */
    providerId: string;
    /** The current review status of the schema submission. */
    status: ProviderSchemaStatus;
    /** The API endpoint where the provider's data can be accessed. */
    apiEndpoint: string;
    /** The provider's GraphQL schema as a string. */
    schema: string;
    /** Metadata defining rules for each field (e.g., consent, ownership). */
    fieldMetadata: object;
}


// IN-MEMORY DATABASES
// Note: In a production environment, these would be replaced with a persistent database.
export const applicationsDB: Application[] = [];
export const providerProfilesDB: ProviderProfile[] = [];
export const providerSchemasDB: ProviderSchema[] = [];


// MOCK/PLACEHOLDER SERVICES

/**
 * @summary Simulates notifying an admin
 * @description In a real system, this would trigger an email, Slack message, or webhook
 * @param {string} message - The notification message for the admin
 */
const notifyAdmin = (message: string): void => {
  console.log(`[Notification] To Admin: ${message}`);
};

/**
 * @summary Simulates updating the `consumer_grants.json` in the Policy Decision Point (PDP)
 * @param {Application} application - The approved application containing the grants
 * @returns {Promise<void>}
 */
const updatePdpPolicy = async (application: Application): Promise<void> => {
  console.log(`[PDP Update] Updating consumer_grants.json for appId: ${application.appId}`);
  const policyPayload = { appId: application.appId, grants: application.requiredFields };
  console.log('[PDP Update] Payload:', JSON.stringify(policyPayload, null, 2));
  await new Promise(resolve => setTimeout(resolve, 500)); // Simulate network delay
  console.log(`[PDP Update] Policy updated successfully for ${application.appId}.`);
};

/**
 * @summary Simulates generating secure API credentials for an application
 * @param {string} appId - The ID of the application
 * @returns {{ apiKey: string, apiSecret: string }} The generated credentials
 */
const generateCredentials = (appId: string): { apiKey: string, apiSecret: string } => {
    console.log(`[Credentials] Generating credentials for ${appId}`);
    const apiKey = `key_${crypto.randomBytes(16).toString('hex')}`;
    const apiSecret = `secret_${crypto.randomBytes(32).toString('hex')}`;
    return { apiKey, apiSecret };
}

/**
 * @summary Simulates updating the `provider_metadata.json` in the Policy Decision Point (PDP)
 * @param {ProviderSchema} schema - The approved provider schema
 * @returns {Promise<void>}
 */
const updatePdpWithProviderMetadata = async (schema: ProviderSchema): Promise<void> => {
    console.log(`[PDP Update] Updating provider_metadata.json for providerId: ${schema.providerId}`);
    const policyPayload = schema.fieldMetadata;
    console.log('[PDP Update] Payload:', JSON.stringify(policyPayload, null, 2));
    await new Promise(resolve => setTimeout(resolve, 500)); // Simulate network delay
    console.log(`[PDP Update] Provider metadata updated successfully for ${schema.providerId}.`);
};


// EXPRESS APPLICATION SETUP
const app = express();
const PORT = 3000;

app.use(express.json()); // Middleware to parse JSON bodies


// CONSUMER & ADMIN API ENDPOINTS

/**
 * @summary [CONSUMER] Submit a new application for data access
 * @route POST /applications
 * @body {string} appId - Unique ID for the consumer's application
 * @body {object} requiredFields - The data fields the application needs
 * @returns {201} The created application object with "pending" status
 * @returns {400} If the request body is missing required fields
 * @returns {409} If an application with the same appId already exists
 */
app.post('/applications', (req: Request, res: Response) => {
  const { appId, requiredFields } = req.body;
  if (!appId || !requiredFields) {
    return res.status(400).json({ error: "Invalid request body: missing 'appId' or 'requiredFields'" });
  }
  if (applicationsDB.some(app => app.appId === appId)) {
      return res.status(409).json({ error: `Application with appId '${appId}' already exists.` });
  }
  const newApplication: Application = { appId, requiredFields, status: 'pending' };
  applicationsDB.push(newApplication);
  console.log('New application received:', newApplication);
  notifyAdmin(`New application '${appId}' requires review.`);
  return res.status(201).json(newApplication);
});

/**
 * @summary [ADMIN] Get all consumer applications, optionally filtering by status
 * @route GET /applications
 * @query {string} [status] - Optional status to filter by (e.g., "pending")
 * @returns {200} An array of application objects
 */
app.get('/applications', (req: Request, res: Response) => {
    const status = req.query.status as ApplicationStatus | undefined;
    if (status) {
        return res.status(200).json(applicationsDB.filter(app => app.status === status));
    }
    return res.status(200).json(applicationsDB);
});

/**
 * @summary [ADMIN] Approve or deny a consumer's application
 * @route POST /applications/:appId/review
 * @param {string} appId - The ID of the application to review
 * @body {string} decision - The decision, must be "approve" or "deny"
 * @returns {200} The updated application object
 * @returns {400} If the decision is invalid
 * @returns {404} If the application is not found
 * @returns {409} If the application is not in a "pending" state
 */
app.post('/applications/:appId/review', async (req: Request, res: Response) => {
  const { appId } = req.params;
  const { decision } = req.body;
  if (decision !== 'approve' && decision !== 'deny') {
    return res.status(400).json({ error: "Invalid decision. Must be 'approve' or 'deny'." });
  }
  const application = applicationsDB.find(app => app.appId === appId);
  if (!application) {
    return res.status(404).json({ error: `Application with appId '${appId}' not found.` });
  }
  if (application.status !== 'pending') {
      return res.status(409).json({ error: `Application is already '${application.status}' and cannot be reviewed again.`})
  }
  if (decision === 'approve') {
    application.status = 'approved';
    console.log(`[Admin] Application ${appId} has been approved.`);
    await updatePdpPolicy(application);
    application.credentials = generateCredentials(appId);
  } else {
    application.status = 'denied';
    console.log(`[Admin] Application ${appId} has been denied.`);
  }
  return res.status(200).json(application);
});


// PROVIDER & ADMIN API ENDPOINTS

/**
 * @summary [PROVIDER] Register a new Data Provider profile
 * @description This is the first step for a new provider. It creates their profile in the system
 * @route POST /providers
 * @body {string} providerName - The official name of the provider
 * @body {string} contactEmail - The primary contact email
 * @body {string} phoneNumber - The contact phone number
 * @body {string} providerType - The type of provider ('government', 'board', 'business')
 * @returns {201} The created provider profile object
 * @returns {400} If the request body is invalid or missing fields
 * @returns {409} If a provider with the same name already exists
 */
app.post('/providers', (req: Request, res: Response) => {
    const { providerName, contactEmail, phoneNumber, providerType } = req.body;

    // Validation
    if (!providerName || !contactEmail || !phoneNumber || !providerType) {
        return res.status(400).json({ error: "Invalid request body: missing one or more required fields." });
    }
    const validTypes: ProviderType[] = ['government', 'board', 'business'];
    if (!validTypes.includes(providerType)) {
        return res.status(400).json({ error: `Invalid providerType. Must be one of: ${validTypes.join(', ')}.` });
    }
    if (providerProfilesDB.some(p => p.providerName.toLowerCase() === providerName.toLowerCase())) {
        return res.status(409).json({ error: `A provider with the name '${providerName}' already exists.` });
    }

    // Creation
    const newProvider: ProviderProfile = {
        providerId: `prov_${crypto.randomBytes(12).toString('hex')}`,
        providerName,
        contactEmail,
        phoneNumber,
        providerType,
        createdAt: new Date().toISOString()
    };

    // Storage & Response
    providerProfilesDB.push(newProvider);
    console.log('New provider registered:', newProvider);
    return res.status(201).json(newProvider);
});


/**
 * @summary [PROVIDER] Submit a new provider schema for approval
 * @route POST /provider-schemas
 * @body {string} providerId - Unique ID for the data provider
 * @body {string} apiEndpoint - The provider's GraphQL API endpoint
 * @body {string} schema - The GraphQL schema as a string
 * @body {object} fieldMetadata - Metadata for each field
 * @returns {201} The created provider schema object with "pending" status
 * @returns {400} If the request body is missing required fields
 * @returns {409} If an active or pending submission for the provider already exists
 */
app.post('/provider-schemas', (req: Request, res: Response) => {
    const { providerId, apiEndpoint, schema, fieldMetadata } = req.body;
    if (!providerId || !apiEndpoint || !schema || !fieldMetadata) {
        return res.status(400).json({ error: "Invalid request body: missing required fields."});
    }
    // Ensure the provider profile exists before allowing schema submission
    if (!providerProfilesDB.some(p => p.providerId === providerId)) {
        return res.status(404).json({ error: `Provider with providerId '${providerId}' not found.`});
    }
    if (providerSchemasDB.some(s => s.providerId === providerId && (s.status === 'pending' || s.status === 'approved'))) {
        return res.status(409).json({ error: `An active or pending schema submission for '${providerId}' already exists.`});
    }
    const newSchema: ProviderSchema = {
        submissionId: `sub_${crypto.randomBytes(12).toString('hex')}`,
        providerId, apiEndpoint, schema, fieldMetadata, status: 'pending'
    };
    providerSchemasDB.push(newSchema);
    console.log('New provider schema received:', newSchema);
    notifyAdmin(`New schema from provider '${providerId}' requires review.`);
    return res.status(201).json(newSchema);
});

/**
 * @summary [PROVIDER & ADMIN] Get a specific provider's schema submission
 * @route GET /provider-schemas/:providerId
 * @param {string} providerId - The ID of the provider
 * @returns {200} The provider schema object
 * @returns {404} If no submission is found for the provider
 */
app.get('/provider-schemas/:providerId', (req: Request, res: Response) => {
    const { providerId } = req.params;
    const schema = providerSchemasDB.find(s => s.providerId === providerId);
    if (!schema) {
        return res.status(404).json({ error: `No schema submission found for provider '${providerId}'.`});
    }
    return res.status(200).json(schema);
});

/**
 * @summary [ADMIN] Get all provider schema submissions, optionally filtering by status
 * @route GET /provider-schemas
 * @query {string} [status] - Optional status to filter by (e.g., "pending")
 * @returns {200} An array of provider schema objects
 */
app.get('/provider-schemas', (req: Request, res: Response) => {
    const status = req.query.status as ProviderSchemaStatus | undefined;
    if (status) {
        return res.status(200).json(providerSchemasDB.filter(s => s.status === status));
    }
    return res.status(200).json(providerSchemasDB);
});

/**
 * @summary [ADMIN] Approve or request changes for a provider's schema submission
 * @route POST /provider-schemas/:submissionId/review
 * @param {string} submissionId - The ID of the schema submission to review
 * @body {string} decision - The decision, must be "approve" or "request_changes"
 * @returns {200} The updated provider schema object
 * @returns {400} If the decision is invalid
 * @returns {404} If the submission is not found
 * @returns {409} If the submission is not in a "pending" state
 */
app.post('/provider-schemas/:submissionId/review', async (req: Request, res: Response) => {
    const { submissionId } = req.params;
    const { decision } = req.body;
    if (decision !== 'approve' && decision !== 'request_changes') {
        return res.status(400).json({ error: "Invalid decision. Must be 'approve' or 'request_changes'." });
    }
    const schema = providerSchemasDB.find(s => s.submissionId === submissionId);
    if (!schema) {
        return res.status(404).json({ error: `Schema submission with ID '${submissionId}' not found.` });
    }
    if (schema.status !== 'pending') {
        return res.status(409).json({ error: `Schema is already '${schema.status}' and cannot be reviewed again.`});
    }
    if (decision === 'approve') {
        schema.status = 'approved';
        console.log(`[Admin] Schema from '${schema.providerId}' has been approved.`);
        await updatePdpWithProviderMetadata(schema);
    } else {
        schema.status = 'changes_required';
        console.log(`[Admin] Changes requested for schema from '${schema.providerId}'.`);
    }
    return res.status(200).json(schema);
});


// SERVER START
// This conditional logic prevents the server from starting during tests
if (process.env.NODE_ENV !== 'test') {
    app.listen(PORT, () => {
      console.log(`Backend server is running on http://localhost:${PORT}`);
    });
}

export default app;

