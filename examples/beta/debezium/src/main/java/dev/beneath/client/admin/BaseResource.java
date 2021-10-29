package dev.beneath.client.admin;

import dev.beneath.client.Connection;

public abstract class BaseResource {
  protected final Connection conn;

  protected BaseResource(Connection conn) {
    this.conn = conn;
  }
}
