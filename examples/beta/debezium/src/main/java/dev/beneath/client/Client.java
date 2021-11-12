package dev.beneath.client;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

import org.apache.avro.generic.GenericRecord;

import dev.beneath.CompileSchemaQuery.CompileSchema;
import dev.beneath.CreateTableMutation.CreateTable;
import dev.beneath.client.admin.AdminClient;
import dev.beneath.client.utils.ProjectIdentifier;
import dev.beneath.client.utils.TableIdentifier;
import dev.beneath.type.CompileSchemaInput;
import dev.beneath.type.CreateTableInput;
import dev.beneath.type.TableSchemaKind;

/**
 * The main class for interacting with Beneath. Data-related features (like
 * defining tables and reading/writing data) are implemented directly on
 * `Client`, while control-plane features (like creating projects) are isolated
 * in the `admin` member.
 * 
 * Args: secret (str): A beneath secret to use for authentication. If not set,
 * uses the ``BENEATH_SECRET`` environment variable, and if that is not set
 * either, uses the secret authenticated in the CLI (stored in ``~/.beneath``).
 * dry (bool): If true, the client will not perform any mutations or writes, but
 * generally perform reads as usual. It's useful for testing.
 * 
 * The exact implication differs for different operations: Some mutations will
 * be mocked, such as creating a table, others will fail with an exception.
 * Write operations log records to the logger instead of transmitting to the
 * server. Reads generally work, but throw an exception when called on mocked
 * resources. write_delay_ms (int): The maximum amount of time to buffer written
 * records before sending a batch write request over the network. Defaults to 1
 * second (1000 ms). Writing records in batches reduces the number of requests,
 * which leads to lower cost (Beneath charges at least 1kb per request).
 */
public class Client {
  public Connection connection;
  public AdminClient adminClient;
  public Boolean dry;
  private Integer startCount;
  private Map<TableIdentifier, Checkpointer> checkpointers;
  private DryWriter dryWriter;
  private Writer writer;

  public Client(String secret, Boolean dry, Integer writeDelayMs) throws Exception {
    this.connection = new Connection(secret);
    this.adminClient = new AdminClient(connection, dry);
    this.dry = dry;

    if (dry) {
      this.dryWriter = new DryWriter(this, writeDelayMs);
    } else {
      this.writer = new Writer(this, writeDelayMs);
    }

    this.startCount = 0;
    this.checkpointers = new HashMap<TableIdentifier, Checkpointer>();
  }

  public Table findTable(String tablePath) throws Exception {
    TableIdentifier identifier = TableIdentifier.fromPath(tablePath);
    Table table = Table.make(this, identifier);
    return table;
  }

  // TODO: Too many method parameters – use custom types, a builder, method
  // overloading, other?
  // TODO: Retention types should be the Java equivalent of Python's "timedelta".
  // Then fix in the CreateTableInputBuilder
  public Table createTable(String tablePath, String schema, String description, Boolean meta, Boolean useIndex,
      Boolean useWarehouse, Integer retention, Integer logRetention, Integer indexRetention, Integer warehouseRetention,
      TableSchemaKind schemaKind, String indexes, Boolean updateIfExists) throws Exception {
    Table table;
    TableIdentifier identifier = TableIdentifier.fromPath(tablePath);
    if (this.dry) {
      CompileSchema data = this.adminClient.tables
          .compileSchema(CompileSchemaInput.builder().schemaKind(schemaKind).schema(schema).build()).get();
      table = Table.makeDry(this, identifier, data.canonicalAvroSchema());
    } else {
      // omitting indexes for now
      CreateTable data = this.adminClient.tables.create(CreateTableInput.builder()
          .organizationName(identifier.organization).projectName(identifier.project).tableName(identifier.table)
          .schemaKind(schemaKind).schema(schema).description(description).meta(meta).useIndex(useIndex)
          .useWarehouse(useWarehouse).logRetentionSeconds(logRetention).indexRetentionSeconds(indexRetention)
          .warehouseRetentionSeconds(warehouseRetention).updateIfExists(updateIfExists).build()).get();
      table = Table.make(this, identifier, data);
    }
    return table;
  }

  // WRITING

  /**
   * Opens the client for writes. Can be called multiple times, but make sure to
   * call ``stop`` correspondingly.
   */
  public void start() throws Exception {
    this.startCount += 1;
    if (this.startCount != 1) {
      return;
    }

    this.connection.ensureConnected();
    if (dry) {
      this.dryWriter.start();
    } else {
      this.writer.start();
    }

    for (Checkpointer checkpointer : this.checkpointers.values()) {
      checkpointer.start();
    }
  }

  /**
   * Closes the client for writes, ensuring buffered writes are flushed. If
   * ``start`` was called multiple times, only the last corresponding call to
   * ``stop`` triggers a flush.
   */
  public void stop() throws Exception {
    if (this.startCount == 0) {
      throw new Exception("Called stop more times than start");
    }

    if (this.startCount == 1) {
      for (Checkpointer checkpointer : this.checkpointers.values()) {
        checkpointer.stop();
      }
      this.writer.stop();
    }

    this.startCount -= 1;
  }

  /**
   * Writes one or more records to ``instance``. By default, writes are buffered
   * for up to ``write_delay_ms`` milliseconds before being transmitted to the
   * server. See the Client constructor for details.
   * 
   * To enabled writes, make sure to call ``start`` on the client (and ``stop``
   * before terminating).
   * 
   * Args: instance (TableInstance): The instance to write to. You can also call
   * ``instance.write`` as a convenience wrapper. records: The records to write.
   * Can be a single record (dict) or a list of records (iterable of dict).
   */
  public void write(TableInstance instance, List<GenericRecord> records) throws Exception {
    if (this.startCount == 0) {
      throw new Exception("Cannot call write because the client is stopped");
    }
    if (dry) {
      this.dryWriter.write(instance, records);
    } else {
      this.writer.write(instance, records);
    }
  }

  /**
   * Forces the client to flush buffered writes without stopping
   */
  public void forceFlush() throws Exception {
    this.writer.forceFlush();
  }

  // CHECKPOINTERS

  /**
   * Returns a checkpointer for the given project. Checkpointers store (small)
   * key-value records useful for maintaining consumer and pipeline state. State
   * is stored in a meta-table called "checkpoints" in the given project.
   * 
   * Args: project_path (str): Path to the project in which to store the
   * checkpointer's state
   * 
   * @throws Exception
   */
  public Checkpointer checkpointer(String projectPath) throws Exception {
    return checkpointer(projectPath, null, "checkpoints", true,
        "Stores checkpointed state for consumers, pipelines, and more");
  }

  /**
   * Returns a checkpointer for the given project. Checkpointers store (small)
   * key-value records useful for maintaining consumer and pipeline state. State
   * is stored in a meta-table called "checkpoints" in the given project.
   * 
   * Args: project_path (str): Path to the project in which to store the
   * checkpointer's state key_prefix (str): If set, any ``get`` or ``set`` call on
   * the checkpointer will prepend the prefix to the key. metatable_name (str):
   * Name of the meta table in which to save checkpointed data metatable_create
   * (bool): If true, the checkpointer will create the checkpoints meta-table if
   * it does not already exists. If false, the checkpointer will throw an
   * exception if the meta-table does not already exist. Defaults to True.
   * metatable_description (str): An optional description to apply to the
   * checkpoints meta-table. Defaults to a sensible description of checkpointing.
   * 
   * @throws Exception
   */
  public Checkpointer checkpointer(String projectPath, String keyPrefix, String metatableName, Boolean metatableCreate,
      String metatableDescription) throws Exception {
    ProjectIdentifier projectIdentifier = ProjectIdentifier.fromPath(projectPath);
    TableIdentifier identifier = new TableIdentifier(projectIdentifier.organization, projectIdentifier.project,
        metatableName);

    if (!this.checkpointers.containsKey(identifier)) {
      Checkpointer checkpointer = new Checkpointer(this, identifier, metatableCreate, metatableDescription);
      this.checkpointers.put(identifier, checkpointer);
      if (this.startCount != 0) {
        checkpointer.start();
      }
    }

    Checkpointer checkpointer = this.checkpointers.get(identifier);
    // TODO: implement PrefixedCheckpointer class
    // if (keyPrefix != null) {
    // checkpointer = PrefixedCheckpointer(checkpointer, keyPrefix);
    // }

    return checkpointer;
  }

}
