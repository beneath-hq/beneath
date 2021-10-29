package dev.beneath.client.admin;

import dev.beneath.client.Connection;

/**
 * AdminClient isolates control-plane features.
 * 
 * Args: connection (Connection): An authenticated connection to Beneath.
 */
public class AdminClient {
  public Connection connection;
  public Tables tables;

  public AdminClient(Connection connection) {
    this.connection = connection;
    this.tables = new Tables(this.connection);

    this.connection.createGraphQlConnection();
  }
}
