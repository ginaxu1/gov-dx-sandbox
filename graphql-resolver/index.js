import { ApolloServer } from '@apollo/server';
import { startStandaloneServer } from '@apollo/server/standalone';
import { ApolloGateway, IntrospectAndCompose, RemoteGraphQLDataSource } from '@apollo/gateway';
import * as dotenv from 'dotenv';

dotenv.config();

const port = process.env.PORT || 4000;

// Define your two Ballerina GraphQL services (subgraphs)
const gateway = new ApolloGateway({
  supergraphSdl: new IntrospectAndCompose({
    subgraphs: [
      // This is your "Registry of Persons" service providing user name
      { name: 'drp', url: process.env.CHOREO_DRP_CONNECTION_SERVICEURL || 'http://localhost:9090/' },

      // This is your "Department of Motor Traffic" service providing license info
      { name: 'dmt', url: process.env.CHOREO_DMT_CONNECTION_SERVICEURL || 'http://localhost:9091/' },
    ],
    // Optional: Set a poll interval to refresh the schema every 10 seconds
    // pollIntervalInMs: 10000, 
  }),
  // Here’s where you attach headers dynamically
  buildService({ name, url }) {
    return new RemoteGraphQLDataSource({
      url,
      willSendRequest({ request }) {
        if (name === 'drp') {
          request.http?.headers.set(
            'Choreo-API-Key',
            process.env.CHOREO_DRP_CONNECTION_CHOREOAPIKEY ?? ''
          );
        }
        if (name === 'dmt') {
          request.http?.headers.set(
            'Choreo-API-Key',
            process.env.CHOREO_DMT_CONNECTION_CHOREOAPIKEY ?? ''
          );
        }
      },
    });
  },
}
);

// Create the server that will expose the single, unified graph
const server = new ApolloServer({
  gateway,
  introspection: true,
});

// Start the server
async function startServer() {
  const { url } = await startStandaloneServer(server, {
    listen: { port },
  });
  console.log(`🚀 Unified Gateway ready at: ${url}`);
}

startServer();