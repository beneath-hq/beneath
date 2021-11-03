package dev.beneath.client;

import java.util.UUID;

import dev.beneath.CreateTableMutation;
import dev.beneath.TableByOrganizationProjectAndNameQuery;
import dev.beneath.CreateTableInstanceMutation.CreateTableInstance;

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
  public static TableInstance make(Client client, Table table, CreateTableInstance adminData) {
    TableInstance instance = new TableInstance();
    instance.client = client;
    instance.table = table;
    instance.setAdminData(adminData);
    return instance;
  }

  // overloaded method to accommodate the
  // "CreateTableMutation.PrimaryTableInstance" type
  public static TableInstance make(Client client, Table table, CreateTableMutation.PrimaryTableInstance adminData) {
    TableInstance instance = new TableInstance();
    instance.client = client;
    instance.table = table;
    instance.setAdminData(adminData);
    return instance;
  }

  // overloaded method to accommodate the
  // "TableByOrganizationProjectAndNameQuery.PrimaryTableInstance" type
  public static TableInstance make(Client client, Table table,
      TableByOrganizationProjectAndNameQuery.PrimaryTableInstance adminData) {
    TableInstance instance = new TableInstance();
    instance.client = client;
    instance.table = table;
    instance.setAdminData(adminData);
    return instance;
  }

  public static TableInstance makeDry(Client client, Table table, Integer version, Boolean makePrimary) {
    TableInstance instance = new TableInstance();
    instance.client = client;
    instance.table = table;
    instance.instanceId = null;
    instance.version = version;
    instance.isPrimary = makePrimary;
    return instance;
  }

  private void setAdminData(CreateTableInstance adminData) {
    this.instanceId = UUID.fromString(adminData.tableInstanceID());
    this.version = adminData.version();
    this.isFinal = adminData.madeFinalOn() != null;
    this.isPrimary = adminData.madePrimaryOn() != null;
  }

  // overloaded method to accommodate the
  // "CreateTableMutation.PrimaryTableInstance" type
  private void setAdminData(CreateTableMutation.PrimaryTableInstance adminData) {
    this.instanceId = UUID.fromString(adminData.tableInstanceID());
    this.version = adminData.version();
    this.isFinal = adminData.madeFinalOn() != null;
    this.isPrimary = adminData.madePrimaryOn() != null;
  }

  // overloaded method to accommodate the
  // "TableByOrganizationProjectAndNameQuery.PrimaryTableInstance" type
  private void setAdminData(TableByOrganizationProjectAndNameQuery.PrimaryTableInstance adminData) {
    this.instanceId = UUID.fromString(adminData.tableInstanceID());
    this.version = adminData.version();
    this.isFinal = adminData.madeFinalOn() != null;
    this.isPrimary = adminData.madePrimaryOn() != null;
  }
}