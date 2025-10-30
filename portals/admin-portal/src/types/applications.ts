export interface SelectedField {
  fieldName: string;
  schemaId: string;}
export interface ApplicationRegistration {
  applicationName: string;
  applicationDescription?: string;
  selectedFields: SelectedField[];
  memberId: string;
}

export interface ApplicationSubmission extends ApplicationRegistration {
  submissionId: string;
  status: 'pending' | 'approved' | 'rejected';
  createdAt: string;
  updatedAt: string;
  review?: string;
}

export interface ApprovedApplication {
  applicationId: string;
  applicationName: string;
  applicationDescription?: string;
  selectedFields: SelectedField[];
  memberId: string;
  version: "active" | "deprecated";
  createdAt: string;
  updatedAt: string;
}
 
// API Response structure
export interface ApplicationSubmissionApiResponse {
  count: number;
  items: ApplicationSubmission[] | null;
}

export interface ApprovedApplicationApiResponse {
  count: number;
  items: ApprovedApplication[] | null;
}