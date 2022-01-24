package dev.beneath.client.admin;

/**
 * GraphQL error
 */
public class GraphQLException extends RuntimeException {
  public GraphQLException(String message) {
    super(message);
  }
}
