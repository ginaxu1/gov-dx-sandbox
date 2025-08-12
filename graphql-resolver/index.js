import { ApolloServer } from '@apollo/server';
import { startStandaloneServer } from '@apollo/server/standalone';
import { ApolloGateway, IntrospectAndCompose } from '@apollo/gateway';

// Define your two Ballerina GraphQL services (subgraphs)
const gateway = new ApolloGateway({
  supergraphSdl: new IntrospectAndCompose({
    subgraphs: [
      // This is your "Registry of Persons" service providing user name
      { name: 'rop', url: 'http://provider-wrapper-drp-service' },

      // This is your "Department of Motor Traffic" service providing license info
      { name: 'dmt', url: 'http://provider-wrapper-dmt-service' },
    ],
    // Optional: Set a poll interval to refresh the schema every 10 seconds
    // pollIntervalInMs: 10000, 
  }),
});

// Create the server that will expose the single, unified graph
const server = new ApolloServer({
  gateway,
});

// Start the server
async function startServer() {
  const { url } = await startStandaloneServer(server, {
    listen: { port: 4000 },
  });
  console.log(`🚀 Unified Gateway ready at: ${url}`);
}

startServer();