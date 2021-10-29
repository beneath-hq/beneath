package dev.beneath.client;

class AuthenticationException extends Exception {
  // Error returned for failed authentication
  AuthenticationException(String message) {
    super(message);
  }
}