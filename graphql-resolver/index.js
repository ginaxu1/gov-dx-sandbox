import { ApolloServer } from '@apollo/server';
import { startStandaloneServer } from '@apollo/server/standalone';
import { ApolloGateway, IntrospectAndCompose } from '@apollo/gateway';
import * as dotenv from 'dotenv';

dotenv.config();

const port = process.env.PORT || 4000;

// Define the Ballerina GraphQL services (subgraphs)
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
        url: process.env.CHOREO_DMT_CONNECTION_SERVICEURL || 'http://localhost:9092/',
        headers: {
          'Choreo-API-Key': process.env.CHOREO_DMT_CONNECTION_APIKEY
        }
      },
    ],
  }),
});

// Create the server that will expose the single, unified graph
const server = new ApolloServer({
  gateway,
  introspection: true
});

// Start the server
async function startServer() {
  const { url } = await startStandaloneServer(server, {
    listen: { port },
  });
  console.log(`Unified Gateway ready at: ${url}`);
}

startServer();