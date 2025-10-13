// types/applications.ts

export interface ApplicationRegistration {
  applicationName: string;
  applicationDescription?: string;
  selectedFields: string[];
}

export interface ApplicationSubmission extends ApplicationRegistration {
  submissionId: string;
  consumerId: string;
  status: 'pending' | 'approved' | 'rejected';
}

export interface ApprovedApplication {
  applicationId: string;
  applicationName: string;
  applicationDescription?: string;
  selectedFields: string[];
  consumerId: string;
  version: "active" | "deprecated";
  createdAt: string;
  updatedAt: string;
}
 
// API Response structure
export interface PendingApplicationApiResponse {
  count: number;
  items: ApplicationSubmission[] | null;
}

export interface ApprovedApplicationApiResponse {
  count: number;
  items: ApprovedApplication[] | null;
}