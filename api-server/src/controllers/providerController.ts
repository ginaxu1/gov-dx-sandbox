
import { Request, Response } from 'express';
import crypto from 'crypto';
import { providerProfilesDB, providerSchemasDB, providerSubmissionsDB } from '../database';
import { notifyAdmin, updatePdpWithProviderMetadata } from '../services';
import { ProviderProfile, ProviderSubmission } from '../types';
import { sendSuccess, sendError } from '../utils/responseHandler';

/**
 * @summary [PROVIDER] Submit a registration request to become a Data Provider
 * @description Creates a temporary submission record that awaits admin approval
 * @route POST /provider-submissions
 */
export const createProviderSubmission = (req: Request, res: Response) => {
    const { providerName, contactEmail, phoneNumber, providerType } = req.body;

    // Validation
    if (!providerName || !contactEmail || !phoneNumber || !providerType) {
        return sendError(res, 'Invalid request body: missing one or more required fields.', 400);
    }
    if (providerSubmissionsDB.some(s => s.providerName.toLowerCase() === providerName.toLowerCase() && s.status === 'pending')) {
        return sendError(res, `A pending submission for '${providerName}' already exists.`, 409);
    }

    // Submission Creation
    const newSubmission: ProviderSubmission = {
        submissionId: `sub_prov_${crypto.randomBytes(12).toString('hex')}`,
        providerName,
        contactEmail,
        phoneNumber,
        providerType,
        status: 'pending',
        createdAt: new Date().toISOString()
    };
    providerSubmissionsDB.push(newSubmission);
    notifyAdmin(`New provider registration for '${providerName}' requires review.`);

    return sendSuccess(res, { submissionId: newSubmission.submissionId }, 'Provider registration submitted for review.', 202);
};  

/**
 * @summary [ADMIN] Get all provider submissions
 */
export const getAllProviderSubmissions = (req: Request, res: Response) => {
    return sendSuccess(res, providerSubmissionsDB, 'Provider submissions retrieved successfully.');
};

/**
 * @summary [ADMIN] Review a provider registration submission
 * @description Approves or rejects a pending provider registration. On approval, creates the permanent provider profile
 * @route POST /provider-submissions/:submissionId/review
 */
export const reviewProviderSubmission = (req: Request, res: Response) => {
    const { submissionId } = req.params;
    const { decision } = req.body;
    const submission = providerSubmissionsDB.find(s => s.submissionId === submissionId);

    if (!submission) {
        return sendError(res, `Provider submission with ID '${submissionId}' not found.`, 404);
    }
    if (submission.status !== 'pending') {
        return sendError(res, `Submission is already '${submission.status}' and cannot be reviewed again.`, 409);
    }

    if (decision === 'approve') {
        const newProvider: ProviderProfile = {
            providerId: `prov_${crypto.randomBytes(12).toString('hex')}`,
            providerName: submission.providerName,
            contactEmail: submission.contactEmail,
            phoneNumber: submission.phoneNumber,
            providerType: submission.providerType,
            approvedAt: new Date().toISOString()
        };
        providerProfilesDB.push(newProvider);
        submission.status = 'approved'; 
        return sendSuccess(res, newProvider, 'Provider submission approved and profile created.');
    } else if (decision === 'reject') {
        submission.status = 'rejected';
        return sendSuccess(res, submission, 'Provider submission has been rejected.');
    } else {
        return sendError(res, "Invalid decision. Must be 'approve' or 'reject'.", 400);
    }
};

/**
 * @summary [PROVIDER] Create a new schema for an existing, approved provider
 * @route POST /providers/:providerId/schemas
 */
export const createSchemaForProvider = (req: Request, res: Response) => {
    const { providerId } = req.params;
    if (!req.body) {
        return sendError(res, "Request body is missing.", 400);
    }
    const { fieldConfigurations } = req.body;

    const provider = providerProfilesDB.find(p => p.providerId === providerId);
    if (!provider) {
        return sendError(res, `Provider with providerId '${providerId}' not found.`, 404);
    }
    if (!fieldConfigurations || Object.keys(fieldConfigurations).length === 0) {
        return sendError(res, "Invalid request body: 'fieldConfigurations' is required.", 400);
    }

    const newSchema: ProviderSchema = {
        submissionId: `sub_schema_${crypto.randomBytes(12).toString('hex')}`,
        providerId,
        status: 'pending',
        fieldConfigurations
    };
    providerSchemasDB.push(newSchema);
    notifyAdmin(`New schema from provider '${providerId}' requires review.`);

    return sendSuccess(res, { providerId: newSchema.providerId }, 'Schema submitted successfully for review.', 201);
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
    const { providerId } = req.params;
    const { decision } = req.body;
    const schema = providerSchemasDB.find(s => s.providerId === providerId && s.status === 'pending');

    if (!schema) {
        return sendError(res, `Schema submission for this provider not found or is no longer pending.`, 404);
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