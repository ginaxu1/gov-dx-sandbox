// services/schemaService.ts
import type { IntrospectionResult, SchemaRegistration } from '../types/graphql';

export class SchemaService {
  private static readonly INTROSPECTION_QUERY = `
    query IntrospectionQuery {
      __schema {
        queryType { name }
        mutationType { name }
        subscriptionType { name }
        types {
          ...FullType
        }
      }
    }

    fragment FullType on __Type {
      kind
      name
      description
      fields(includeDeprecated: true) {
        name
        description
        args {
          ...InputValue
        }
        type {
          ...TypeRef
        }
      }
      inputFields {
        ...InputValue
      }
      interfaces {
        ...TypeRef
      }
      enumValues(includeDeprecated: true) {
        name
        description
        isDeprecated
        deprecationReason
      }
      possibleTypes {
        ...TypeRef
      }
    }

    fragment InputValue on __InputValue {
      name
      description
      type { ...TypeRef }
      defaultValue
    }

    fragment TypeRef on __Type {
      kind
      name
      ofType {
        kind
        name
        ofType {
          kind
          name
          ofType {
            kind
            name
            ofType {
              kind
              name
              ofType {
                kind
                name
                ofType {
                  kind
                  name
                  ofType {
                    kind
                    name
                  }
                }
              }
            }
          }
        }
      }
    }
  `;

  static async fetchSchemaFromEndpoint(endpoint: string): Promise<IntrospectionResult> {
    try {
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          query: this.INTROSPECTION_QUERY,
        }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result = await response.json();
      
      if (result.errors) {
        throw new Error(`GraphQL errors: ${result.errors.map((e: any) => e.message).join(', ')}`);
      }

      return result;
    } catch (error) {
      throw new Error(`Failed to fetch schema: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  static async parseIntrospectionJSON(file: File): Promise<IntrospectionResult> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = (event) => {
        try {
          const result = JSON.parse(event.target?.result as string);
          
          // Validate basic structure
          if (!result.data?.__schema?.types) {
            throw new Error('Invalid introspection result format');
          }
          
          resolve(result);
        } catch (error) {
          reject(new Error(`Failed to parse JSON: ${error instanceof Error ? error.message : 'Invalid JSON format'}`));
        }
      };
      reader.onerror = () => reject(new Error('Failed to read file'));
      reader.readAsText(file);
    });
  }

  static async parseSDL(sdl: string): Promise<IntrospectionResult> {
    try {
      if (!sdl.trim()) {
        throw new Error('Empty SDL string');
      }

      // Dynamically import to avoid adding graphql to initial bundle if not needed
      const { buildSchema, getIntrospectionQuery, graphql } = await import('graphql');

      const schema = buildSchema(sdl);
      const introspectionQuery = getIntrospectionQuery();

      const result = await graphql({
        schema,
        source: introspectionQuery,
      });

      if (result.errors?.length) {
        throw new Error(`SDL introspection errors: ${result.errors.map(e => e.message).join(', ')}`);
      }

      if (!result.data || !(result as any).data.__schema?.types) {
        throw new Error('Invalid introspection result from SDL');
      }

      return result as unknown as IntrospectionResult;
    } catch (error) {
      throw new Error(`Failed to parse SDL: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  static async registerSchema(providerId: string, registration: SchemaRegistration): Promise<void> {
    const baseUrl = import.meta.env.VITE_BASE_PATH || '';
    console.log('Registering schema at:', `${baseUrl}providers/${providerId}/schemas`);
    try {
      const response = await fetch(`${baseUrl}providers/${providerId}/schemas`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(registration),
      });

      if (!response.ok) {
        throw new Error(`Registration failed! status: ${response.status}`);
      }
    } catch (error) {
      throw new Error(`Failed to register schema: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
}