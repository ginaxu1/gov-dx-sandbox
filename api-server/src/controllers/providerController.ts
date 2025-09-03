import { Request, Response } from "express";
import crypto from "crypto";
import { providerProfilesDB, providerSchemasDB } from "../database";
import {
  ProviderProfile,
  ProviderSchema,
  ProviderType,
  ProviderSchemaStatus,
} from "../types";
import { notifyAdmin, updatePdpWithProviderMetadata } from "../services";

/**
 * @summary [PROVIDER] Register a new Data Provider profile
 */
export const createProviderProfile = (req: Request, res: Response) => {
  const { providerName, contactEmail, phoneNumber, providerType } = req.body;
  if (!providerName || !contactEmail || !phoneNumber || !providerType) {
    return res
      .status(400)
      .json({
        error: "Invalid request body: missing one or more required fields.",
      });
  }
  const validTypes: ProviderType[] = ["government", "board", "business"];
  if (!validTypes.includes(providerType)) {
    return res
      .status(400)
      .json({
        error: `Invalid providerType. Must be one of: ${validTypes.join(", ")}.`,
      });
  }
  if (
    providerProfilesDB.some(
      (p) => p.providerName.toLowerCase() === providerName.toLowerCase(),
    )
  ) {
    return res
      .status(409)
      .json({
        error: `A provider with the name '${providerName}' already exists.`,
      });
  }
  const newProvider: ProviderProfile = {
    providerId: `prov_${crypto.randomBytes(12).toString("hex")}`,
    providerName,
    contactEmail,
    phoneNumber,
    providerType,
    createdAt: new Date().toISOString(),
  };
  providerProfilesDB.push(newProvider);
  return res.status(201).json(newProvider);
};

/**
 * @summary [PROVIDER] Submit a new provider schema for approval
 */
export const createProviderSchema = (req: Request, res: Response) => {
  const { providerId, apiEndpoint, schema, fieldMetadata } = req.body;
  if (!providerId || !apiEndpoint || !schema || !fieldMetadata) {
    return res
      .status(400)
      .json({ error: "Invalid request body: missing required fields." });
  }
  if (!providerProfilesDB.some((p) => p.providerId === providerId)) {
    return res
      .status(404)
      .json({ error: `Provider with providerId '${providerId}' not found.` });
  }
  if (
    providerSchemasDB.some(
      (s) =>
        s.providerId === providerId &&
        (s.status === "pending" || s.status === "approved"),
    )
  ) {
    return res
      .status(409)
      .json({
        error: `An active or pending schema submission for '${providerId}' already exists.`,
      });
  }
  const newSchema: ProviderSchema = {
    submissionId: `sub_${crypto.randomBytes(12).toString("hex")}`,
    providerId,
    apiEndpoint,
    schema,
    fieldMetadata,
    status: "pending",
  };
  providerSchemasDB.push(newSchema);
  notifyAdmin(`New schema from provider '${providerId}' requires review.`);
  return res.status(201).json(newSchema);
};

/**
 * @summary [PROVIDER & ADMIN] Get a specific provider's schema submission by providerId
 */
export const getProviderSchema = (req: Request, res: Response) => {
  const { providerId } = req.params;
  const schema = providerSchemasDB.find((s) => s.providerId === providerId);
  if (!schema) {
    return res
      .status(404)
      .json({
        error: `No schema submission found for provider '${providerId}'.`,
      });
  }
  return res.status(200).json(schema);
};

/**
 * @summary [ADMIN] Get all provider schema submissions
 */
export const getAllProviderSchemas = (req: Request, res: Response) => {
  const status = req.query.status as ProviderSchemaStatus | undefined;
  if (status) {
    return res
      .status(200)
      .json(providerSchemasDB.filter((s) => s.status === status));
  }
  return res.status(200).json(providerSchemasDB);
};

/**
 * @summary [ADMIN] Approve or request changes for a provider's schema submission
 */
export const reviewProviderSchema = async (req: Request, res: Response) => {
  const { submissionId } = req.params;
  const { decision } = req.body;
  if (decision !== "approve" && decision !== "request_changes") {
    return res
      .status(400)
      .json({
        error: "Invalid decision. Must be 'approve' or 'request_changes'.",
      });
  }
  const schema = providerSchemasDB.find((s) => s.submissionId === submissionId);
  if (!schema) {
    return res
      .status(404)
      .json({
        error: `Schema submission with ID '${submissionId}' not found.`,
      });
  }
  if (schema.status !== "pending") {
    return res
      .status(409)
      .json({
        error: `Schema is already '${schema.status}' and cannot be reviewed again.`,
      });
  }
  if (decision === "approve") {
    schema.status = "approved";
    await updatePdpWithProviderMetadata(schema);
  } else {
    schema.status = "changes_required";
  }
  return res.status(200).json(schema);
};
