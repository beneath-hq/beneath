package dev.beneath.client;

import java.util.UUID;

import dev.beneath.CreateTableInstanceMutation.CreateTableInstance;
import dev.beneath.CreateTableMutation.CreateTable;
import dev.beneath.TableByOrganizationProjectAndNameQuery.TableByOrganizationProjectAndName;
import dev.beneath.client.utils.TableIdentifier;
import dev.beneath.type.CreateTableInstanceInput;

/**
 * Represents a data-plane connection to a table. To find or create a table, see
 * :class:`beneath.Client`.
 * 
 * Use it to get a TableInstance, which you can query, replay, subscribe and
 * write to. Learn more about tables and instances at
 * https://about.beneath.dev/docs/concepts/tables/.
 */
public class Table {
  public UUID tableId;
  public Schema schema;
  public TableInstance primaryInstance;
  public Boolean useLog;
  public Boolean useIndex;
  public Boolean useWarehouse;
  private Client client;
  public TableIdentifier identifier;

  Table() {
  }

  // INITIALIZATION

  // TODO: Review these methods. Too much method overloading? Should I follow the
  // "Builder" pattern?
  public static Table make(Client client, TableIdentifier identifier) throws Exception {
    Table table = new Table();
    table.client = client;
    table.identifier = identifier;
    TableByOrganizationProjectAndName adminData = table.loadAdminData();
    table.tableId = UUID.fromString(adminData.tableID());
    table.schema = new Schema(adminData.avroSchema());
    if (adminData.primaryTableInstance() != null) {
      table.primaryInstance = TableInstance.make(client, table, adminData.primaryTableInstance());
    }
    table.useLog = adminData.useLog();
    table.useIndex = adminData.useIndex();
    table.useWarehouse = adminData.useWarehouse();
    return table;
  }

  // overloaded method to include "adminData"
  public static Table make(Client client, TableIdentifier identifier, TableByOrganizationProjectAndName adminData)
      throws Exception {
    Table table = new Table();
    table.client = client;
    table.identifier = identifier;
    table.tableId = UUID.fromString(adminData.tableID());
    table.schema = new Schema(adminData.avroSchema());
    if (adminData.primaryTableInstance() != null) {
      table.primaryInstance = TableInstance.make(client, table, adminData.primaryTableInstance());
    }
    table.useLog = adminData.useLog();
    table.useIndex = adminData.useIndex();
    table.useWarehouse = adminData.useWarehouse();
    return table;
  }

  // overload method to accommodate "CreateTable" type
  public static Table make(Client client, TableIdentifier identifier, CreateTable adminData) throws Exception {
    Table table = new Table();
    table.client = client;
    table.identifier = identifier;
    table.tableId = UUID.fromString(adminData.tableID());
    table.schema = new Schema(adminData.avroSchema());
    if (adminData.primaryTableInstance() != null) {
      table.primaryInstance = TableInstance.make(client, table, adminData.primaryTableInstance());
    }
    table.useLog = adminData.useLog();
    table.useIndex = adminData.useIndex();
    table.useWarehouse = adminData.useWarehouse();
    return table;
  }

  public static Table makeDry(Client client, TableIdentifier identifier, String avroSchema) throws Exception {
    Table table = new Table();
    table.client = client;
    table.identifier = identifier;
    table.tableId = null;
    table.schema = new Schema(avroSchema);
    table.primaryInstance = table.createInstance(0, true, true);
    table.useLog = true;
    table.useIndex = true;
    table.useWarehouse = true;
    return table;
  }

  private TableByOrganizationProjectAndName loadAdminData() throws Exception {
    return this.client.adminClient.tables
        .findByOrganizationProjectAndName(this.identifier.organization, this.identifier.project, this.identifier.table)
        .get();
  }

  // STATE

  @Override
  public String toString() {
    return String.format("<beneath.table.Table(\"%s/%s\")>", Config.BENEATH_FRONTEND_HOST, this.identifier.toString());
  }

  // INSTANCES

  public TableInstance createInstance(Integer version, Boolean makePrimary, Boolean updateIfExists) throws Exception {
    TableInstance instance;
    // handle real and dry cases
    if (this.tableId != null) {
      CreateTableInstance adminData = this.client.adminClient.tables
          .createInstance(CreateTableInstanceInput.builder().tableID(this.tableId.toString()).version(version)
              .makePrimary(makePrimary).updateIfExists(updateIfExists).build())
          .get();
      instance = TableInstance.make(this.client, this, adminData);
    } else {
      instance = TableInstance.makeDry(this.client, this, version, makePrimary);
    }
    if (makePrimary) {
      this.primaryInstance = instance;
    }
    return instance;
  }
}
