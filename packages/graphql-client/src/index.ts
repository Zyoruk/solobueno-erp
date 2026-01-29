/**
 * @solobueno/graphql - GraphQL Client
 *
 * Type-safe GraphQL client generated from backend schema.
 * Provides queries, mutations, and subscriptions for all operations.
 *
 * @packageDocumentation
 */

// Re-export types for convenience
export * from '@solobueno/types';

// TODO: Generated GraphQL operations will be exported here
// export { useMenuQuery, useMenuMutation } from './generated/menu';
// export { useOrdersQuery, useCreateOrderMutation } from './generated/orders';
// export { useTablesQuery, useUpdateTableMutation } from './generated/tables';

export const GRAPHQL_VERSION = '0.0.1';

/**
 * Placeholder GraphQL client configuration
 */
export interface GraphQLClientConfig {
  endpoint: string;
  headers?: Record<string, string>;
}

export function createGraphQLClient(_config: GraphQLClientConfig) {
  // TODO: Implement GraphQL client
  return {
    query: async () => {
      throw new Error('GraphQL client not implemented');
    },
    mutate: async () => {
      throw new Error('GraphQL client not implemented');
    },
  };
}
