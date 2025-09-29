// types/applications.ts

export interface ApplicationRegistration {
  name: string;
  description?: string;
  selectedFields: string[];
  callback_url?: string;
  homepage_url?: string;
}

export interface ApplicationSubmission extends ApplicationRegistration {
  submissionId: string;
  created_at: string;
  status: 'pending' | 'approved' | 'rejected';
  consumerId: string;
}

export interface ApprovedApplication {
  applicationId: string;
  name: string;
  description?: string;
  selectedFields: string[];
  callback_url?: string;
  homepage_url?: string;
  client_id?: string;
  client_secret?: string;
  version: "Active" | "Deprecated";
  created_at: string;
  consumerId: string;
}

export interface ApplicationConfiguration {
  fieldAccess: 'required' | 'optional' | 'restricted' | '';
  description: string;
  isUserDefinedField: boolean;
}

export interface ApplicationApiResponse<T> {
  count: number;
  items: T[] | null;
}