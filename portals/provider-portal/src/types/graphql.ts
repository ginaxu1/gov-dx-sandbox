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
  source: 'authorative' | 'fallback' | 'other';
  isOwner: boolean;
  description: string;
  isQueryType: boolean; // Is Field Defined Inside a Query Type
  isUserDefinedTypeField: boolean; // Is Field is a User Defined Type 
}

export interface SchemaRegistration {
  sdl: string;
}