// graphql-resolver/index.js
import { ApolloServer } from '@apollo/server';
import { startStandaloneServer } from '@apollo/server/standalone';
import { ApolloGateway, IntrospectAndCompose } from '@apollo/gateway';
import { GraphQLError } from 'graphql'; // To throw GraphQL-specific errors
import * as dotenv from 'dotenv';

dotenv.config();

const port = process.env.PORT || 4000;
// Configure the URL for your Policy Governance Go service
const POLICY_GOVERNANCE_SERVICE_URL = process.env.POLICY_GOVERNANCE_SERVICE_URL || 'http://localhost:8081/evaluate-policy';

/**
 * A simplified function to extract requested field names from a GraphQL operation.
 * NOTE: For a production-grade gateway, this function would need to be much more
 * sophisticated to handle nested fields, fragments, aliases, and to accurately
 * determine `subgraphName` and `typeName` based on the supergraph schema.
 * This example makes basic inferences based on known fields for demonstration.
 * @param {object} operationDefinition The parsed GraphQL operation definition.
 * @returns {Array<object>} An array of requested field objects.
 */
function extractRequestedFields(operationDefinition) {
  const fields = [];
  if (operationDefinition && operationDefinition.selectionSet && operationDefinition.selectionSet.selections) {
    operationDefinition.selectionSet.selections.forEach(selection => {
      if (selection.kind === 'Field') {
        // This is a basic extraction. Full AST traversal is complex.
        // We'll augment with known subgraph/type info below for the policy service.
        fields.push({
          fieldName: selection.name.value,
          // Placeholder values; actual values derived from schema mapping in real setup
          subgraphName: 'unknown',
          typeName: 'unknown',
          classification: 'ALLOW', // Default client assumption
          context: {}
        });
      }
      // TODO: Handle fragments, inline fragments, and nested fields for a complete solution
    });
  }
  return fields;
}

// Apollo Server Plugin for Policy Governance integration
const policyGovernancePlugin = {
  // This hook is called after the GraphQL operation is parsed and validated.
  async requestDidStart(requestContext) {
    return {
      async didResolveOperation({ request, document }) {
        const operationDefinition = document.definitions.find(
          (def) => def.kind === 'OperationDefinition' && def.operation === 'query' // Focusing on queries
          // TODO: Extend to mutations if policy checks are needed for them
        );

        if (!operationDefinition) {
          // If no query operation is found (e.g., introspection query, or mutation not handled), skip.
          return;
        }

        // Extract raw requested fields from the client's GraphQL query
        const rawRequestedFields = extractRequestedFields(operationDefinition);

        if (rawRequestedFields.length === 0) {
          console.log('Policy Governance: No fields identified for policy check.');
          return; // No fields to check, let the request proceed
        }

        // Augment requestedFields with more accurate `subgraphName` and `typeName`.
        // In a real system, this mapping would come from a sophisticated supergraph schema parser
        // or dynamic schema awareness within the gateway. For this example, we infer based on field names.
        const augmentedFields = rawRequestedFields.map(field => {
            // Infer subgraph and type for fields known to be in your example database
            if (['engineNumber', 'ownerNic', 'registrationNumber', 'vehicleClass'].includes(field.fieldName)) {
                return { ...field, subgraphName: 'dmt', typeName: 'VehicleInfo' };
            }
            if (['id'].includes(field.fieldName) && field.subgraphName === 'dmt') { // Distinguish 'id'
                return { ...field, subgraphName: 'dmt', typeName: 'VehicleInfo' }; // Assuming VehicleInfo for 'dmt' id
            }
            if (field.fieldName === 'photo') {
                return { ...field, subgraphName: 'drp', typeName: 'PersonData' };
            }
            if (field.fieldName === 'licenseNumber') {
                 return { ...field, subgraphName: 'dmt', typeName: 'DriverLicense' };
            }
            // Add other inferences as needed for your schema
            return { ...field, subgraphName: 'unknown', typeName: 'unknown' }; // Default if no inference
        });

        // Construct the payload for the Policy Governance service
        const policyRequestPayload = {
          // Assuming consumerId might be passed in context from a previous middleware (e.g., HTTP headers)
          consumerId: requestContext.contextValue?.consumerId || 'default-consumer-id',
          requestedFields: augmentedFields,
        };

        try {
          console.log('Policy Governance: Calling service with payload:', JSON.stringify(policyRequestPayload, null, 2));

          const response = await fetch(POLICY_GOVERNANCE_SERVICE_URL, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify(policyRequestPayload),
          });

          if (!response.ok) {
            console.error(`Policy Governance Service returned HTTP error: ${response.status} ${response.statusText}`);
            throw new GraphQLError('Policy service error: Could not determine access. Please try again later.');
          }

          const policyResponse = await response.json();
          console.log('Policy Governance: Service response:', JSON.stringify(policyResponse, null, 2));

          // Check if overall consent is required for any field
          if (policyResponse.overallConsentRequired) {
            const consentRequiredDetails = policyResponse.accessScopes
              .filter(scope => scope.consentRequired)
              .map(scope => `${scope.subgraphName}.${scope.typeName}.${scope.fieldName} (Consent Type: ${scope.consentType})`)
              .join(', ');

            throw new GraphQLError(
              `Access requires consent for fields: ${consentRequiredDetails}. Please obtain consent and retry the request.`,
              {
                extensions: {
                  code: 'CONSENT_REQUIRED',
                  details: policyResponse.accessScopes.filter(scope => scope.consentRequired),
                },
              }
            );
          }

          // Check for explicitly DENIED fields
          const deniedFields = policyResponse.accessScopes
            .filter(scope => scope.resolvedClassification === 'DENIED')
            .map(scope => `${scope.subgraphName}.${scope.typeName}.${scope.fieldName}`);

          if (deniedFields.length > 0) {
            throw new GraphQLError(
              `Access denied for fields: ${deniedFields.join(', ')}.`,
              {
                extensions: {
                  code: 'ACCESS_DENIED',
                  details: policyResponse.accessScopes.filter(scope => scope.resolvedClassification === 'DENIED'),
                },
              }
            );
          }

          // If no consent is required and no fields are denied, the request proceeds normally.
          console.log('Policy Governance: All checks passed. Request proceeding to subgraphs.');

        } catch (error) {
          // Re-throw GraphQLError instances directly
          if (error instanceof GraphQLError) {
            throw error;
          }
          // Wrap other errors (e.g., network errors from fetch) in a generic GraphQL error
          console.error('Policy Governance: An unexpected error occurred during policy check:', error);
          throw new GraphQLError('An internal server error occurred during policy evaluation.');
        }
      },
    };
  },
};

// Define the Ballerina GraphQL services (subgraphs) for the Gateway
const gateway = new ApolloGateway({
  supergraphSdl: new IntrospectAndCompose({
    subgraphs: [
      {
        name: 'drp',
        url: process.env.CHOREO_DRP_CONNECTION_SERVICEURL || 'http://localhost:9091/',
        headers: {
          'Choreo-API-Key': process.env.CHOREO_DRP_CONNECTION_APIKEY
        }
      },
      {
        name: 'dmt',
        url: process.env.CHOREO_DMT_CONNECTION_SERVICEURL || 'http://localhost:9090/',
        headers: {
          'Choreo-API-Key': process.env.CHOREO_DMT_CONNECTION_APIKEY
        }
      },
      // Add more subgraphs as needed
    ],
  }),
});

// Create the Apollo Server instance
const server = new ApolloServer({
  gateway,
  introspection: true, // Enable introspection for tools like Apollo Sandbox
  plugins: [policyGovernancePlugin], // Register our custom policy governance plugin
});

// Start the Apollo Server
async function startServer() {
  const { url } = await startStandaloneServer(server, {
    listen: { port },
    // Example: Pass consumer ID from request headers into the context
    // This allows the policy governance plugin to access it.
    context: async ({ req }) => ({
      consumerId: req.headers['x-consumer-id'] || 'anonymous',
    }),
  });
  console.log(`🚀 Unified Gateway ready at: ${url}`);
  console.log(`Policy Governance Service URL: ${POLICY_GOVERNANCE_SERVICE_URL}`);
}

startServer();
