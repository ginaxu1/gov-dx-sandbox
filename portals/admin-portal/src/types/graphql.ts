// types/graphql.ts
export interface GraphQLField {
  name: string;
  type: {
    kind: string;
    name?: string;
    ofType?: any;
  };
  description?: string;
  args?: GraphQLInputValue[];
}

export interface GraphQLInputValue {
  name: string;
  type: {
    kind: string;
    name?: string;
    ofType?: any;
  };
  description?: string;
  defaultValue?: any;
}

export interface GraphQLType {
  kind: string;
  name: string;
  description?: string;
  fields?: GraphQLField[];
  inputFields?: GraphQLInputValue[];
  interfaces?: GraphQLType[];
  possibleTypes?: GraphQLType[];
  enumValues?: Array<{
    name: string;
    description?: string;
    isDeprecated?: boolean;
    deprecationReason?: string;
  }>;
}

export interface IntrospectionResult {
  data: {
    __schema: {
      queryType: { name: string };
      mutationType?: { name: string };
      subscriptionType?: { name: string };
      types: GraphQLType[];
    };
  };
}

export interface FieldConfiguration {
  accessControlType: 'public' | 'restricted' | ''; // '' indicates not set
  source: 'authoritative' | 'fallback' | 'other' | ''; // '' indicates not set
  isOwner: boolean | null;
  owner: string; // Owner Identifier
  description: string;
  isQueryType: boolean; // Is Field Defined Inside a Query Type
  isUserDefinedTypeField: boolean; // Is this field a User Defined Type field
}

export interface SchemaRegistration {
  schemaName: string;
  schemaDescription?: string;
  sdl: string;
  previousSchemaId: string | null;
  schemaEndpoint: string;
}

export interface SchemaSubmission extends SchemaRegistration {
  submissionId: string;
  status: 'pending' | 'approved' | 'rejected';
  providerId: string;
  createdAt: string; // Note: API uses createdAt, not created_at
  updatedAt: string;
}

export interface ApprovedSchema {
  schemaId: string;
  schemaName: string;
  schemaDescription?: string;
  sdl: string;
  schemaEndpoint: string;
  version: "active" | "deprecated";
  providerId: string;
  createdAt: string;
  updatedAt: string;
}

// API Response structure
export interface PendingSchemaApiResponse {
  count: number;
  items: SchemaSubmission[] | null;
}

export interface ApprovedSchemaApiResponse {
  count: number;
  items: ApprovedSchema[] | null;
}