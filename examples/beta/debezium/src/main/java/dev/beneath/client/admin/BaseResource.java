package dev.beneath.client.admin;

import dev.beneath.client.Connection;

public abstract class BaseResource {
  protected final Connection conn;
  protected final Boolean dry;

  protected BaseResource(Connection conn, Boolean dry) {
    this.conn = conn;
    this.dry = dry;
  }

  protected void beforeMutation() throws Exception {
    if (this.dry) {
      throw new Exception("Cannot run mutation on a client where dry=True");
    }
  }
}
