package dev.beneath.client;

import java.util.UUID;

import dev.beneath.TableByOrganizationProjectAndNameQuery.TableByOrganizationProjectAndName;
import dev.beneath.client.utils.TableIdentifier;

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
  private TableIdentifier identifier;

  Table() {
  }

  // TODO: this should probably follow the "Builder" pattern
  public static Table make(Client client, TableIdentifier identifier, TableByOrganizationProjectAndName adminData)
      throws Exception {
    Table table = new Table();
    table.client = client;
    table.identifier = identifier;
    if (adminData == null) {
      adminData = table.loadAdminData();
    }
    table.tableId = UUID.fromString(adminData.tableID());
    table.schema = new Schema(adminData.avroSchema()); // TODO: create schema class
    if (adminData.primaryTableInstance() != null) {
      table.primaryInstance = TableInstance.make(client, table, adminData.primaryTableInstance());
    }
    table.useLog = adminData.useLog();
    table.useIndex = adminData.useIndex();
    table.useWarehouse = adminData.useWarehouse();
    return table;
  }

  private TableByOrganizationProjectAndName loadAdminData() throws Exception {
    return this.client.adminClient.tables
        .findByOrganizationProjectAndName(this.identifier.organization, this.identifier.project, this.identifier.table)
        .get();
  }
}
