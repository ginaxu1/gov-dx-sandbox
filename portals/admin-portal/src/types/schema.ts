export interface SchemaRegistration {
  schemaName: string;
  schemaDescription?: string;
  sdl: string;
  previousSchemaId: string | null;
  schemaEndpoint: string;
  memberId: string;
}

export interface SchemaSubmission extends SchemaRegistration {
  submissionId: string;
  status: 'pending' | 'approved' | 'rejected';
  createdAt: string;
  updatedAt: string;
  review?: string;
}

export interface ApprovedSchema {
  schemaId: string;
  schemaName: string;
  schemaDescription?: string;
  sdl: string;
  schemaEndpoint: string;
  version: "active" | "deprecated";
  createdAt: string;
  updatedAt: string;
  memberId: string;
}

// API Response structure
export interface SchemaSubmissionApiResponse {
  count: number;
  items: SchemaSubmission[] | null;
}

export interface ApprovedSchemaApiResponse {
  count: number;
  items: ApprovedSchema[] | null;
}