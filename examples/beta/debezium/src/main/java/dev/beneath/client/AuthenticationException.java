package dev.beneath.client;

/**
 * Error returned for failed authentication
 */
class AuthenticationException extends RuntimeException {
  AuthenticationException(String message) {
    super(message);
  }
}