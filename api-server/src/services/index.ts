import crypto from "crypto";
import { Application, ProviderSchema } from "../types";

// Simulates notifying an admin
export const notifyAdmin = (message: string): void => {
  console.log(`[Notification] To Admin: ${message}`);
};

// Simulates updating the `consumer_grants.json` in the PDP
export const updatePdpPolicy = async (
  application: Application,
): Promise<void> => {
  console.log(
    `[PDP Update] Updating consumer_grants.json for appId: ${application.appId}`,
  );
  await new Promise((resolve) => setTimeout(resolve, 500)); // Simulate network delay
  console.log(
    `[PDP Update] Policy updated successfully for ${application.appId}.`,
  );
};

// Simulates generating secure API credentials for an application
export const generateCredentials = (
  appId: string,
): { apiKey: string; apiSecret: string } => {
  console.log(`[Credentials] Generating credentials for ${appId}`);
  const apiKey = `key_${crypto.randomBytes(16).toString("hex")}`;
  const apiSecret = `secret_${crypto.randomBytes(32).toString("hex")}`;
  return { apiKey, apiSecret };
};

// Simulates updating the `provider_metadata.json` in the PDP
export const updatePdpWithProviderMetadata = async (
  schema: ProviderSchema,
): Promise<void> => {
  console.log(
    `[PDP Update] Updating provider_metadata.json for providerId: ${schema.providerId}`,
  );
  await new Promise((resolve) => setTimeout(resolve, 500)); // Simulate network delay
  console.log(
    `[PDP Update] Provider metadata updated successfully for ${schema.providerId}.`,
  );
};
