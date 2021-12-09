package dev.beneath.client;

import java.util.UUID;
import java.util.concurrent.ExecutionException;

import com.google.protobuf.ByteString;

import dev.beneath.CreateTableInstanceMutation.CreateTableInstance;
import dev.beneath.CreateTableMutation.CreateTable;
import dev.beneath.TableByOrganizationProjectAndNameQuery.TableByOrganizationProjectAndName;
import dev.beneath.client.utils.TableIdentifier;
import dev.beneath.type.CreateTableInstanceInput;

/**
 * Represents a data-plane connection to a table. To find or create a table, see
 * :class:`beneath.BeneathClient`.
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
  private BeneathClient client;
  public TableIdentifier identifier;

  Table() {
  }

  // INITIALIZATION

  // TODO: Review these methods. Too much method overloading? Should I follow the
  // "Builder" pattern?
  public static Table make(BeneathClient client, TableIdentifier identifier) {
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
  public static Table make(BeneathClient client, TableIdentifier identifier,
      TableByOrganizationProjectAndName adminData) {
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
  public static Table make(BeneathClient client, TableIdentifier identifier, CreateTable adminData) {
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

  public static Table makeDry(BeneathClient client, TableIdentifier identifier, String avroSchema) {
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

  private TableByOrganizationProjectAndName loadAdminData() {
    TableByOrganizationProjectAndName table;
    try {
      table = this.client.adminClient.tables.findByOrganizationProjectAndName(this.identifier.organization,
          this.identifier.project, this.identifier.table).get();
    } catch (InterruptedException | ExecutionException e) {
      throw new RuntimeException(e);
    }
    return table;
  }

  // STATE

  @Override
  public String toString() {
    return String.format("<beneath.table.Table(\"%s/%s\")>", Config.BENEATH_FRONTEND_HOST, this.identifier.toString());
  }

  // INSTANCES

  public TableInstance createInstance(Integer version, Boolean makePrimary, Boolean updateIfExists) {
    TableInstance instance;
    // handle real and dry cases
    if (this.tableId != null) {
      CreateTableInstance adminData;
      try {
        adminData = this.client.adminClient.tables
            .createInstance(CreateTableInstanceInput.builder().tableID(this.tableId.toString()).version(version)
                .makePrimary(makePrimary).updateIfExists(updateIfExists).build())
            .get();
      } catch (InterruptedException | ExecutionException e) {
        throw new RuntimeException(e);
      }
      instance = TableInstance.make(this.client, this, adminData);
    } else {
      instance = TableInstance.makeDry(this.client, this, version, makePrimary);
    }
    if (makePrimary) {
      this.primaryInstance = instance;
    }
    return instance;
  }

  // CURSORS

  /**
   * Restores a cursor previously obtained by querying one of the table's
   * instances. You must provide the cursor bytes, which can be found as
   * properties of the Cursor object.
   */
  public Cursor restoreCursor(ByteString replayCursor, ByteString changesCursor) {
    return new Cursor(this.client.connection, this.schema, replayCursor, changesCursor);
  }
}
