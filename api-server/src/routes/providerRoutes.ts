import express from 'express';
import {
   createProviderSubmission,
   reviewProviderSubmission,
   createSchemaForProvider,
   getSchemaForProvider,
   getAllSchemas,
   reviewSchema,
   getAllProviderSubmissions
} from '../controllers/providerController';

const router = express.Router();

// [PROVIDER] Submit a registration to become a Data Provider
router.post('/provider-submissions', createProviderSubmission);

// [ADMIN] Get all provider submissions
router.get('/provider-submissions', getAllProviderSubmissions);

// [ADMIN] Review a provider registration submission
router.post('/provider-submissions/:submissionId/review', reviewProviderSubmission);

// [PROVIDER] Create a new schema for an existing, approved provider
router.post('/providers/:providerId/schemas', createSchemaForProvider);

// [PROVIDER & ADMIN] Get a specific provider's schema
router.get('/providers/:providerId/schema', getSchemaForProvider);

// [ADMIN] Get all provider schemas
router.get('/provider-schemas', getAllSchemas);

// [ADMIN] Approve or request changes for a provider's schema
router.post('/provider-schemas/:providerId/review', reviewSchema);

export default router;
