import 'dotenv/config'; 
import { ApolloServer } from '@apollo/server';
import { startStandaloneServer } from '@apollo/server/standalone';
import { ApolloGateway, IntrospectAndCompose } from '@apollo/gateway';

// Initialize the gateway and point it to the running subgraphs
const gateway = new ApolloGateway({
  supergraphSdl: new IntrospectAndCompose({
    subgraphs: [
      { name: 'rop', url: process.env.ROP_PROVIDER_URL },
      { name: 'dmv', url: process.env.DMV_PROVIDER_URL },
    ],
  }),
});

const server = new ApolloServer({
  gateway,
});

// Start the server
const { url } = await startStandaloneServer(server, {
  listen: { port: 8080 }, // Listen on the port Choreo expects
});
console.log(`TEST: Gateway ready at ${url}`);


// TODO: Jul 24 > node index.js
// Error: A valid schema couldn't be composed. The following composition errors were found:
    //     Non-shareable field "User.id" is resolved from multiple subgraphs: it is resolved from subgraphs "dmv" and "rop" and defined as non-shareable in subgraph "rop"
    // at IntrospectAndCompose.createSupergraphFromSubgraphList 