import { ConsentStatus } from "../constants/consentStatus";

interface ConsentField {
  fieldName: string;
  schemaID: string;
  displayName?: string;
  description?: string;
  owner: string;
}

export interface ConsentRecord {
  consentId?: string;
  appId: string;
  appName?: string;
  ownerId: string;
  ownerEmail: string;
  status: ConsentStatus;
  type: string;
  createdAt: string;
  updatedAt: string;
  fields: ConsentField[];
  redirectUrl?: string;
}