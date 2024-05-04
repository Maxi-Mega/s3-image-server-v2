import { ApolloClient, createHttpLink, InMemoryCache } from "@apollo/client/core";
import { resolveBackendURL } from "@/composables/url";

// HTTP connection to the API
const httpLink = createHttpLink({
  // You should use an absolute URL here
  uri: resolveBackendURL("/api/graphql"),
});

// Cache implementation
const cache = new InMemoryCache();

// Create the apollo client
export const apolloClient = new ApolloClient({
  link: httpLink,
  cache,
});
