import { Request, Response } from 'express';
import crypto from 'crypto';
import { providerProfilesDB, providerSchemasDB } from '../database';
import { notifyAdmin, updatePdpWithProviderMetadata } from '../services';
import { ProviderProfile, ProviderSchema, ProviderType, FieldConfigurations } from '../types';

/**
 * @summary [PROVIDER] Register a new Data Provider profile
 */
export const createProviderProfile = (req: Request, res: Response) => {
    const { providerName, contactEmail, phoneNumber, providerType } = req.body;

    // Validation Logic
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

    // Creation Logic
    const newProvider: ProviderProfile = {
        providerId: `prov_${crypto.randomBytes(12).toString('hex')}`,
        providerName, contactEmail, phoneNumber, providerType,
        createdAt: new Date().toISOString()
    };
    providerProfilesDB.push(newProvider);
    return res.status(201).json(newProvider);
};


/**
 * @summary [PROVIDER] Register a new, detailed data schema for an existing provider
 * @description Handles the complex schema submission with nested field configurations.
 * @route POST /providers/:providerId/register_schema
 */
export const registerSchema = (req: Request, res: Response) => {
    const { providerId } = req.params;
    const { fieldConfigurations, schemaInput } = req.body;

    // Validation Logic
    if (!providerProfilesDB.some(p => p.providerId === providerId)) {
        return res.status(404).json({ error: `Provider with providerId '${providerId}' not found.` });
    }

    if (!fieldConfigurations || Object.keys(fieldConfigurations).length === 0) {
        return res.status(400).json({ error: "Invalid request body: must include 'fieldConfigurations'." });
    }

    // Detailed Payload Validation
    try {
        for (const typeName in fieldConfigurations) {
            const fields = fieldConfigurations[typeName];
            for (const fieldName in fields) {
                const field = fields[fieldName];
                if (typeof field.isOwner !== 'boolean' || !field.source) {
                    throw new Error(`Invalid configuration for field '${typeName}.${fieldName}'. 'isOwner' must be a boolean and 'source' must be provided.`);
                }
            }
        }
    } catch (error: any) {
        return res.status(400).json({ error: error.message });
    }
    
    if (providerSchemasDB.some(s => s.providerId === providerId && (s.status === 'pending' || s.status === 'approved'))) {
        return res.status(409).json({ error: `An active or pending schema submission for '${providerId}' already exists.` });
    }

    // Creation Logic
    const newSchema: ProviderSchema = {
        submissionId: `sub_${crypto.randomBytes(12).toString('hex')}`,
        providerId,
        status: 'pending',
        schemaInput,
        fieldConfigurations,
    };

    providerSchemasDB.push(newSchema);
    notifyAdmin(`New schema from provider '${providerId}' requires review.`);
    return res.status(201).json(newSchema);
};

/**
 * @summary [PROVIDER & ADMIN] Get a specific provider's schema submission
 */
export const getSchemaForProvider = (req: Request, res: Response) => {
    const { providerId } = req.params;
    const schema = providerSchemasDB.find(s => s.providerId === providerId);
    if (!schema) {
        return res.status(404).json({ error: `No schema submission found for provider '${providerId}'.` });
    }
    return res.status(200).json(schema);
};

/**
 * @summary [ADMIN] Get all provider schema submissions
 */
export const getAllSchemas = (req: Request, res: Response) => {
    const { status } = req.query;
    if (status) {
        return res.status(200).json(providerSchemasDB.filter(s => s.status === String(status)));
    }
    return res.status(200).json(providerSchemasDB);
};

/**
 * @summary [ADMIN] Approve or request changes for a provider's schema
 */
export const reviewSchema = async (req: Request, res: Response) => {
    const { submissionId } = req.params;
    const { decision } = req.body;
    const schema = providerSchemasDB.find(s => s.submissionId === submissionId);

    if (!schema) {
        return res.status(404).json({ error: `Schema submission with ID '${submissionId}' not found.` });
    }
    if (schema.status !== 'pending') {
        return res.status(409).json({ error: `Schema is already '${schema.status}' and cannot be reviewed again.`});
    }

    if (decision === 'approve') {
        schema.status = 'approved';
        await updatePdpWithProviderMetadata(schema);
    } else if (decision === 'request_changes') {
        schema.status = 'changes_required';
    } else {
        return res.status(400).json({ error: "Invalid decision. Must be 'approve' or 'request_changes'." });
    }
    return res.status(200).json(schema);
};

