import express from 'express';
import {
    createProviderProfile,
    registerSchema,
    getSchemaForProvider,
    getAllSchemas,
    reviewSchema
} from '../controllers/providerController';

const router = express.Router();

// [PROVIDER] Register a new Data Provider profile
router.post('/providers', createProviderProfile);

// [PROVIDER] Register a new, detailed data schema for an existing provider
router.post('/providers/:providerId/register_schema', registerSchema);

// [PROVIDER & ADMIN] Get a specific provider's schema submission
router.get('/providers/:providerId/schema', getSchemaForProvider);

// [ADMIN] Get all provider schema submissions
router.get('/provider-schemas', getAllSchemas);

// [ADMIN] Approve or request changes for a provider's schema
router.post('/provider-schemas/:submissionId/review', reviewSchema);


export default router;