import { Request, Response } from 'express';
import crypto from 'crypto';
import { providerProfilesDB, providerSchemasDB } from '../database';
import { notifyAdmin, updatePdpWithProviderMetadata } from '../services';
import { ProviderProfile, ProviderSchema, FieldConfigurations, ProviderType } from '../types';
import { sendSuccess, sendError } from '../utils/responseHandler';

/**
 * @summary [PROVIDER] Register a new Data Provider profile
 */
export const createProviderProfile = (req: Request, res: Response) => {
    const { providerName, contactEmail, phoneNumber, providerType } = req.body;

    // Validation Logic
    if (!providerName || !contactEmail || !phoneNumber || !providerType) {
        return sendError(res, 'Invalid request body: missing one or more required fields.', 400);
    }
    const validTypes: ProviderType[] = ['government', 'board', 'business'];
    if (!validTypes.includes(providerType)) {
        return sendError(res, `Invalid providerType. Must be one of: ${validTypes.join(', ')}.`, 400);
    }
    if (providerProfilesDB.some(p => p.providerName.toLowerCase() === providerName.toLowerCase())) {
        return sendError(res, `A provider with the name '${providerName}' already exists.`, 409);
    }

    // Creation Logic
    const newProvider: ProviderProfile = {
        providerId: `prov_${crypto.randomBytes(12).toString('hex')}`,
        providerName, contactEmail, phoneNumber, providerType,
        createdAt: new Date().toISOString()
    };
    providerProfilesDB.push(newProvider);
    
    return sendSuccess(res, newProvider, 'Provider profile created successfully.', 201);
};


/**
 * @summary [PROVIDER] Register a new, detailed data schema for an existing provider
 */
export const registerSchema = (req: Request, res: Response) => {
     if (!req.body) {
        return sendError(res, "Request body is missing.", 400);
    }
    
    const { providerId } = req.params;
    const { fieldConfigurations } = req.body;

    // Validation Logic
    if (!providerProfilesDB.some(p => p.providerId === providerId)) {
        return sendError(res, `Provider with providerId '${providerId}' not found.`, 404);
    }
    if (!fieldConfigurations || Object.keys(fieldConfigurations).length === 0) {
        return sendError(res, "Invalid request body: 'fieldConfigurations' is required and cannot be empty.", 400);
    }
    if (providerSchemasDB.some(s => s.providerId === providerId && (s.status === 'pending' || s.status === 'approved'))) {
        return sendError(res, `An active or pending schema submission for '${providerId}' already exists.`, 409);
    }

    // Creation Logic
    const newSchema: ProviderSchema = {
        submissionId: `sub_${crypto.randomBytes(12).toString('hex')}`,
        providerId,
        status: 'pending',
        fieldConfigurations
    };
    providerSchemasDB.push(newSchema);
    notifyAdmin(`New schema from provider '${providerId}' requires review.`);

    return sendSuccess(res, newSchema, 'Schema submitted successfully for review.', 201);
};

/**
 * @summary [PROVIDER & ADMIN] Get a specific provider's schema submission
 */
export const getSchemaForProvider = (req: Request, res: Response) => {
    const { providerId } = req.params;
    const schema = providerSchemasDB.find(s => s.providerId === providerId);
    if (!schema) {
        return sendError(res, `No schema submission found for provider '${providerId}'.`, 404);
    }
    return sendSuccess(res, schema);
};

/**
 * @summary [ADMIN] Get all provider schema submissions
 */
export const getAllSchemas = (req: Request, res: Response) => {
    const { status } = req.query;
    if (status) {
        const filteredSchemas = providerSchemasDB.filter(s => s.status === String(status));
        return sendSuccess(res, filteredSchemas);
    }
    return sendSuccess(res, providerSchemasDB);
};

/**
 * @summary [ADMIN] Approve or request changes for a provider's schema
 */
export const reviewSchema = async (req: Request, res: Response) => {
    const { submissionId } = req.params;
    const { decision } = req.body;
    const schema = providerSchemasDB.find(s => s.submissionId === submissionId);

    if (!schema) {
        return sendError(res, `Schema submission with ID '${submissionId}' not found.`, 404);
    }
    if (schema.status !== 'pending') {
        return sendError(res, `Schema is already '${schema.status}' and cannot be reviewed again.`, 409);
    }

    if (decision === 'approve') {
        schema.status = 'approved';
        await updatePdpWithProviderMetadata(schema);
        return sendSuccess(res, schema, 'Schema has been approved.');
    } else if (decision === 'request_changes') {
        schema.status = 'changes_required';
        return sendSuccess(res, schema, 'Schema has been marked as requiring changes.');
    } else {
        return sendError(res, "Invalid decision. Must be 'approve' or 'request_changes'.", 400);
    }
};

