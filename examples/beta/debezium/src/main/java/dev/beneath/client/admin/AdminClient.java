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

  public AdminClient(Connection connection, Boolean dry) {
    this.connection = connection;
    this.tables = new Tables(this.connection, dry);

    this.connection.createGraphQlConnection();
  }
}
