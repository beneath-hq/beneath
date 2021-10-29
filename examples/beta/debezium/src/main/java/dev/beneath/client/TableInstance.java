package dev.beneath.client;

import java.util.UUID;

import dev.beneath.TableByOrganizationProjectAndNameQuery.PrimaryTableInstance;

/**
 * Represents an instance of a Table, i.e. a specific version that you can
 * query/subscribe/write to. Learn more about instances at
 * https://about.beneath.dev/docs/concepts/tables/.
 */
public class TableInstance {
  public Table table;
  public UUID instanceId;
  public Boolean isFinal;
  public Boolean isPrimary;
  public Integer version;
  private Client client;

  public TableInstance() {
  }

  // TODO: this should probably follow the "Builder" pattern
  public static TableInstance make(Client client, Table table, PrimaryTableInstance adminData) {
    TableInstance instance = new TableInstance();
    instance.client = client;
    instance.table = table;
    instance.setAdminData(adminData);
    return instance;
  }

  private void setAdminData(PrimaryTableInstance adminData) {
    this.instanceId = UUID.fromString(adminData.tableInstanceID());
    this.version = adminData.version();
    this.isFinal = adminData.madeFinalOn() != null;
    this.isPrimary = adminData.madePrimaryOn() != null;
  }
}
