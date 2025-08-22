// src/apollo.js
import { ApolloClient, InMemoryCache, createHttpLink } from '@apollo/client';

const GRAPHQL_API_URI = import.meta.env.VITE_GRAPHQL_API_URI || 'http://localhost:4000/';

const client = new ApolloClient({
  link: new HttpLink({
    uri: GRAPHQL_API_URI,
    headers: {
      // TODO: Add Passport organization's API key here for prod. For sandbox, placeholder is fine
      'X-API-KEY': 'your-passport-api-key',
    },
  }),
  cache: new InMemoryCache(),
});

export default client;
