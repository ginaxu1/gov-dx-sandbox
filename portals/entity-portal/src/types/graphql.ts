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
  sdl: string;
  previous_schema_id: string | null;
  schema_endpoint: string;
}

export interface SchemaSubmission extends SchemaRegistration {
  submissionId: string;
  created_at: string;
  status: 'pending' | 'approved' | 'rejected';
  providerId: string;
}

export interface ApprovedSchema {
  schemaId: string;
  sdl: string;
  schema_endpoint: string;
  version: "Active" | "Deprecated";
  created_at: string;
  providerId: string;
}