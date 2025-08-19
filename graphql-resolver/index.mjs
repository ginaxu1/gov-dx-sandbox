import { ApolloServer } from '@apollo/server';
import { ApolloGateway, IntrospectAndCompose } from '@apollo/gateway';
import { startStandaloneServer } from '@apollo/server/standalone';
import { GraphQLError } from 'graphql';
import { Kind } from 'graphql/language/index.js';
import * as dotenv from 'dotenv'; 

dotenv.config();
const POLICY_GOVERNANCE_SERVICE_URL = process.env.POLICY_GOVERNANCE_SERVICE_URL || 'http://localhost:8081/evaluate-policy';
const port = process.env.PORT || 4000;

// --- Helper to map GraphQL Type Names to their respective Subgraph Names ---
// This mapping is crucial for federated schemas, as the 'info' object (or AST)
// doesn't directly provide subgraph information. This map must be kept in sync
// with your Ballerina subgraph definitions.
const TYPE_TO_SUBGRAPH_MAP = {
  // DRP Subgraph Types (from drp/types.bal)
  PersonData: "drp", // The root entity for DRP
  PersonInfo: "drp", // Fields spread from PersonInfo into PersonData
  CardInfo: "drp",
  LostCardReplacementInfo: "drp",
  CitizenshipInfo: "drp",
  ParentInfo: "drp",
  Gender: "drp", // Enums also belong to a subgraph context
  CardStatus: "drp",
  CivilStatus: "drp",
  CitizenshipType: "drp",
  // Added 'person' (lowercase p) type mapping, as indicated by the error log.
  // This assumes 'person' is a distinct type name in your schema for PersonData.
  person: "drp", // Added this line for the 'person' type

  // DMT Subgraph Types (from dmt/types.bal)
  VehicleClass: "dmt",
  VehicleInfo: "dmt",
  DriverLicense: "dmt",
  // Added 'Vehicle' (capital V) type mapping.
  Vehicle: "dmt",
  // Added 'vehicle' (lowercase v) type mapping, assuming Query.vehicle might return a type with this name.
  // This helps resolve 'unknown.vehicle.vehicleInfoById' error if 'vehicle' is indeed a type.
  vehicle: "dmt",
  // Note: 'PersonData' is also present in DMT as an @subgraph:Entity,
  // indicating that DMT contributes fields (vehicles, license) to it.
  // The primary source of the 'PersonData' *type definition* for core fields
  // is considered DRP. For fields contributed by DMT, their specific types
  // (VehicleInfo, DriverLicense) define the DMT subgraph context.
};

/**
 * Recursively collects all requested fields from a GraphQL operation's selection set.
 * This function processes Field, InlineFragment, and FragmentSpread nodes to build
 * the `RequestedField` payload for the policy governance service.
 *
 * @param {Array<object>} selections The 'selections' array from a SelectionSet (GraphQL AST node).
 * @param {object} schema The GraphQLSchema object from `requestContext.schema`.
 * @param {string} currentTypeName The name of the GraphQL type currently being processed (e.g., 'Query', 'PersonData').
 * @param {Set<string>} collectedFieldIdentifiers A Set to track unique fields as "subgraph.Type.field" to avoid duplicates.
 * @param {Array<object>} requestedFieldsList An array to accumulate the structured RequestedField objects.
 * @param {object} fragments Map of named fragments (e.g., `document.fragments`).
 */
function collectRequestedFieldsRecursive(selections, schema, currentTypeName, collectedFieldIdentifiers, requestedFieldsList, fragments) {
  if (!selections) {
    return;
  }

  selections.forEach(selection => {
    if (selection.kind === Kind.FIELD) {
      const fieldName = selection.name.value;

      // Skip __typename and other introspection fields that don't represent actual data.
      if (fieldName.startsWith('__')) {
        return;
      }

      const typeDef = schema.getType(currentTypeName);
      // Ensure we are dealing with an object type that has fields (e.g., not a scalar or enum)
      if (!typeDef || typeof typeDef.getFields !== 'function') {
        return; // Cannot get fields for this type (e.g., scalar, enum)
      }

      const fieldDef = typeDef.getFields()[fieldName];
      if (!fieldDef) {
        // This can happen if the schema validation failed or a field is dynamically added/removed.
        console.warn(`[Field Extractor] Field '${fieldName}' not found on type '${currentTypeName}'. Skipping.`);
        return;
      }

      // Determine the concrete return type name for nested traversal.
      // We need to unwrap NonNullType and ListType to get to the base type name.
      let returnType = fieldDef.type;
      while (returnType && returnType.ofType) {
        returnType = returnType.ofType;
      }
      const returnTypeName = returnType ? returnType.name : null;


      // Determine the subgraph name for the current field.
      let subgraphName;
      if (currentTypeName === "Query") {
        // Explicitly map root query fields to their respective subgraphs based on what they return.
        // Adjusted to include 'person' and 'getPersonByNic' if they are root query fields.
        if (fieldName === 'getPersonDataByNic' || fieldName === 'overallConsentStatus' || fieldName === 'person' || fieldName === 'getPersonByNic') {
          subgraphName = "drp";
        } else if (fieldName === 'vehicle' || fieldName === 'getVehicleById' || fieldName === 'vehicleInfoById') {
          subgraphName = "dmt";
        } else {
          // Fallback for any other root query fields not explicitly mapped.
          // You might want to map these explicitly if they come from a specific subgraph.
          subgraphName = "unknown_query_root_field";
        }
      } else if (currentTypeName === "PersonData") {
        // Special handling for federated 'PersonData': 'vehicles' and 'license' are from DMT.
        if (fieldName === 'vehicles' || fieldName === 'license') {
          subgraphName = "dmt";
        } else {
          // Other fields on 'PersonData' itself (like 'fullName', 'photo') are from DRP.
          subgraphName = TYPE_TO_SUBGRAPH_MAP[currentTypeName] || "unknown_drp_field";
        }
      } else {
        // For other types, use the direct map lookup.
        subgraphName = TYPE_TO_SUBGRAPH_MAP[currentTypeName] || "unknown";
      }

      // Create a unique identifier for the field to prevent duplicates.
      const fieldIdentifier = `${subgraphName}.${currentTypeName}.${fieldName}`;

      if (!collectedFieldIdentifiers.has(fieldIdentifier)) {
        collectedFieldIdentifiers.add(fieldIdentifier);
        requestedFieldsList.push({
          subgraphName: subgraphName,
          typeName: currentTypeName,
          fieldName: fieldName,
          context: {} // Context can be populated with additional information if needed by policy service
        });
      }

      // Recursively process nested selections if the current field has them
      // and its return type is an object type (i.e., it can have sub-fields).
      if (selection.selectionSet && returnTypeName && schema.getType(returnTypeName) && typeof schema.getType(returnTypeName).getFields === 'function') {
        collectRequestedFieldsRecursive(selection.selectionSet.selections, schema, returnTypeName, collectedFieldIdentifiers, requestedFieldsList, fragments);
      }

    } else if (selection.kind === Kind.INLINE_FRAGMENT) {
      // If it's an inline fragment, extract the type it applies to and recurse.
      const fragmentTypeName = selection.typeCondition.name.value;
      collectRequestedFieldsRecursive(selection.selectionSet.selections, schema, fragmentTypeName, collectedFieldIdentifiers, requestedFieldsList, fragments);
    } else if (selection.kind === Kind.FRAGMENT_SPREAD) {
      // If it's a fragment spread, find the fragment definition and recurse.
      const fragmentName = selection.name.value;
      const fragmentDef = fragments[fragmentName];
      if (fragmentDef) {
        const fragmentTypeName = fragmentDef.typeCondition.name.value;
        collectRequestedFieldsRecursive(fragmentDef.selectionSet.selections, schema, fragmentTypeName, collectedFieldIdentifiers, requestedFieldsList, fragments);
      }
    }
  });
}

/**
 * Extracts all requested fields from a GraphQL document, providing their
 * subgraph, type, and field names. This is used to build the payload for
 * the policy governance service.
 *
 * @param {object} document The parsed GraphQL document (AST).
 * @param {object} schema The GraphQLSchema object.
 * @param {string} operationName The name of the operation being executed (or 'default').
 * @returns {Array<object>} An array of { subgraphName, typeName, fieldName, context } objects.
 */
function extractFieldsWithSchema(document, schema, operationName) {
  const collectedFieldIdentifiers = new Set();
  const requestedFieldsList = [];
  const fragments = document.definitions.filter(def => def.kind === Kind.FRAGMENT_DEFINITION)
                                      .reduce((acc, frag) => ({ ...acc, [frag.name.value]: frag }), {});

  let operationDefinition;
  if (operationName) {
      operationDefinition = document.definitions.find(def =>
          def.kind === Kind.OPERATION_DEFINITION && def.name?.value === operationName
      );
  } else {
      // If no operationName, find the first operation definition
      operationDefinition = document.definitions.find(def =>
          def.kind === Kind.OPERATION_DEFINITION
      );
  }

  if (operationDefinition && operationDefinition.selectionSet) {
    let rootType;
    if (operationDefinition.operation === 'query') {
      rootType = schema.getQueryType();
    } else if (operationDefinition.operation === 'mutation') {
      rootType = schema.getMutationType();
    } else if (operationDefinition.operation === 'subscription') {
      rootType = schema.getSubscriptionType();
    }

    if (rootType) {
      collectRequestedFieldsRecursive(
        operationDefinition.selectionSet.selections,
        schema,
        rootType.name,
        collectedFieldIdentifiers,
        requestedFieldsList,
        fragments
      );
    }
  }
  return requestedFieldsList;
}

// Apollo Server Plugin for Policy Governance integration
const policyGovernancePlugin = {
  // This hook is called after the GraphQL operation is parsed and validated.
  async requestDidStart(requestContext) {
    return {
      async didResolveOperation({ request, document }) {
        let operationDefinition;
        if (request.operationName) {
            operationDefinition = document.definitions.find(def =>
                def.kind === Kind.OPERATION_DEFINITION && def.name?.value === request.operationName
            );
        } else {
            // If no explicit operation name, find the first operation definition
            operationDefinition = document.definitions.find(def =>
                def.kind === Kind.OPERATION_DEFINITION
            );
        }

        if (!operationDefinition) {
          console.log('Policy Governance: No valid operation definition found for policy check. Document definitions:', JSON.stringify(document.definitions, null, 2));
          // If no operation, there's nothing to check policies on, so return early.
          return;
        }

        const actualOperationName = operationDefinition.name?.value || 'default';

        // Extract all requested fields from the GraphQL query document
        const augmentedFields = extractFieldsWithSchema(document, requestContext.schema, actualOperationName);
        console.log('Policy Governance: Augmented fields for policy check:', JSON.stringify(augmentedFields, null, 2));

        if (augmentedFields.length === 0) {
          console.log('Policy Governance: No fields identified for policy check.');
          // If no fields are requested, no policies need to be checked.
          return;
        }

        // Construct the payload for the policy governance service
        const policyRequestPayload = {
          consumerId: requestContext.contextValue?.consumerId || 'anonymous-consumer', // Use 'anonymous-consumer' as default
          requestedFields: augmentedFields,
        };

        try {
          console.log('Policy Governance: Calling service with payload:', JSON.stringify(policyRequestPayload, null, 2));

          // Make the HTTP POST request to the policy governance service
          const response = await fetch(POLICY_GOVERNANCE_SERVICE_URL, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              // Add any other necessary headers, e.g., Authorization token
              // 'Authorization': requestContext.request.http.headers.get('authorization'),
            },
            body: JSON.stringify(policyRequestPayload),
          });

          if (!response.ok) {
            console.error(`Policy Governance Service returned HTTP error: ${response.status} ${response.statusText}`);
            // Throw a GraphQL error to the client if the service call fails
            throw new GraphQLError('Policy service error: Could not determine access due to an upstream issue.');
          }

          const policyResponse = await response.json();
          console.log('Policy Governance: Service response:', JSON.stringify(policyResponse, null, 2));

          // Store the policy response in the context so downstream resolvers can access it
          // This is crucial for the `createPolicyEnforcedResolver` in subgraphs to work.
          requestContext.contextValue.policyGovernanceResult = policyResponse;

          // --- Handle DENIED fields ---
          // If any field is explicitly DENIED, the Gateway plugin will still halt the *entire* query.
          // This is the primary gatekeeping mechanism.
          const deniedFields = policyResponse.accessScopes
            .filter(scope => scope.resolvedClassification === 'DENY')
            .map(scope => `${scope.subgraphName}.${scope.typeName}.${scope.fieldName}`);

          if (deniedFields.length > 0) {
            console.warn(`Policy Governance: Access DENIED for fields: ${deniedFields.join(', ')}. Halting request.`);
            throw new GraphQLError(
              `Access denied for fields: ${deniedFields.join(', ')}.`,
              {
                extensions: {
                  code: 'ACCESS_DENIED',
                  details: policyResponse.accessScopes.filter(scope => scope.resolvedClassification === 'DENY'),
                },
              }
            );
          }

          // --- Handle consent-required fields (do nothing for now as per new requirement) ---
          // If fields require consent, the request will now proceed without interruption at the Gateway.
          // The `policyGovernanceResult` in the context will carry this information, and it's
          // expected that a future Consent Engine or the individual field resolvers will manage this.
          const consentRequiredFields = policyResponse.accessScopes.filter(scope =>
            ['ALLOW_PROVIDER_CONSENT', 'ALLOW_CITIZEN_CONSENT', 'ALLOW_CONSENT'].includes(scope.resolvedClassification)
          );

          if (consentRequiredFields.length > 0) {
            const consentRequiredDetails = consentRequiredFields
              .map(scope => `${scope.subgraphName}.${scope.typeName}.${scope.fieldName} (Classification: ${scope.resolvedClassification})`)
              .join(', ');
            console.log(`Policy Governance: Fields requiring consent detected. Request will proceed to subgraphs without blocking here. Consent handling to be implemented downstream: ${consentRequiredDetails}`);
            // NO THROW: The request will continue processing.
          }

          console.log('Policy Governance: All checks passed (no denials, consent deferred). Request proceeding to subgraphs.');

        } catch (error) {
          // Re-throw GraphQLError instances directly
          if (error instanceof GraphQLError) {
            throw error;
          }
          // Catch and wrap any other unexpected errors during policy evaluation
          console.error('Policy Governance: An unexpected error occurred during policy check:', error);
          throw new GraphQLError('An internal server error occurred during policy evaluation.');
        }
      },
    };
  },
};

// Configure the Apollo Gateway to compose your subgraphs
const gateway = new ApolloGateway({
  supergraphSdl: new IntrospectAndCompose({
    subgraphs: [
      {
        name: 'drp',
        url: process.env.CHOREO_DRP_CONNECTION_SERVICEURL || 'http://localhost:9091/',
        // TODO: include headers for API keys or other authentication
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
    ],
  }),
});

// Create the Apollo Server instance
const server = new ApolloServer({
  gateway,
  // Enable introspection for GraphQL Playground/tools
  introspection: true,
  // Attach the custom policy governance plugin
  plugins: [policyGovernancePlugin],
});

// Start the Apollo Server
async function startServer() {
  const { url } = await startStandaloneServer(server, {
    listen: { port },
    context: async ({ req }) => ({
      // Extract consumerId from headers and add it to the context
      consumerId: req.headers['x-consumer-id'] || 'anonymous',
      // The policyGovernanceResult will be added to this context by the plugin
    }),
  });
  console.log(`Unified Gateway ready at: ${url}`);
  console.log(`Policy Governance Service URL: ${POLICY_GOVERNANCE_SERVICE_URL}`);
}

startServer();
