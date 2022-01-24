package dev.beneath.client;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutionException;

import org.apache.avro.generic.GenericRecord;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import dev.beneath.client.CompileSchemaQuery.CompileSchema;
import dev.beneath.client.CreateProjectMutation.CreateProject;
import dev.beneath.client.CreateTableMutation.CreateTable;
import dev.beneath.client.OrganizationByNameQuery.OrganizationByName;
import dev.beneath.client.admin.AdminClient;
import dev.beneath.client.type.CompileSchemaInput;
import dev.beneath.client.type.CreateProjectInput;
import dev.beneath.client.type.CreateTableInput;
import dev.beneath.client.type.TableSchemaKind;
import dev.beneath.client.utils.ProjectIdentifier;
import dev.beneath.client.utils.SubscriptionIdentifier;
import dev.beneath.client.utils.TableIdentifier;
import dev.beneath.client.utils.Utils;

/**
 * The main class for interacting with Beneath. Data-related features (like
 * defining tables and reading/writing data) are implemented directly on
 * `BeneathClient`, while control-plane features (like creating projects) are
 * isolated in the `admin` member.
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
public class BeneathClient {
  public Connection connection;
  public AdminClient adminClient;
  public Boolean dry;
  private Integer startCount;
  private Map<TableIdentifier, Checkpointer> checkpointers;
  private Map<SubscriptionIdentifier, Consumer> consumers;
  private DryWriter dryWriter;
  private Writer writer;

  private static final Logger LOGGER = LoggerFactory.getLogger(BeneathClient.class);

  public BeneathClient(String secret, Boolean dry, Integer writeDelayMs) {
    this.connection = new Connection(secret);
    this.adminClient = new AdminClient(connection, dry);
    this.dry = dry;

    if (dry) {
      this.dryWriter = new DryWriter(writeDelayMs);
    } else {
      this.writer = new Writer(this, writeDelayMs);
    }

    this.startCount = 0;
    this.checkpointers = new HashMap<TableIdentifier, Checkpointer>();
    this.consumers = new HashMap<SubscriptionIdentifier, Consumer>();
  }

  // FINDING AND STAGING PROJECTS

  public CreateProject createProject(String organizationName, String projectName, Boolean public_, String description,
      String site) {
    OrganizationByName organization;
    try {
      organization = this.adminClient.organizations
          .findByName(organizationName).get();
    } catch (InterruptedException | ExecutionException e) {
      throw new RuntimeException(e);
    }

    CreateProject project;
    try {
      project = this.adminClient.projects
          .create(CreateProjectInput.builder().organizationID(organization.organizationID())
              .projectName(projectName).public_(public_)
              .description(description)
              .site(site).build())
          .get();
    } catch (InterruptedException | ExecutionException e) {
      throw new RuntimeException(e);
    }

    return project;
  }

  public void stageProject(String organizationName, String projectName, Boolean public_, String description,
      String site) {
    try {
      this.adminClient.projects
          .findByOrganizationAndName(organizationName, projectName).get();
      LOGGER.info("Using existing project '{}/{}'", organizationName, projectName);
    } catch (Exception e) {
      if (e.getMessage().contains(String.format("Project %s/%s not found", organizationName, projectName))) {
        LOGGER.info("Project '{}/{}' not found", organizationName, projectName);
        this.createProject(organizationName, projectName, public_, description, site);
        LOGGER.info("Created a new project '{}/{}'", organizationName, projectName);
        return;
      }
      throw new RuntimeException(e);
    }
  }

  // FINDING AND STAGING TABLES

  public Table findTable(String tablePath) {
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
      TableSchemaKind schemaKind, String indexes, Boolean updateIfExists) {
    Table table;
    TableIdentifier identifier = TableIdentifier.fromPath(tablePath);
    if (this.dry) {
      CompileSchema data;
      try {
        data = this.adminClient.tables
            .compileSchema(CompileSchemaInput.builder().schemaKind(schemaKind).schema(schema).build()).get();
      } catch (InterruptedException | ExecutionException e) {
        throw new RuntimeException(e);
      }
      table = Table.makeDry(this, identifier, data.canonicalAvroSchema());
    } else {
      // omitting indexes for now
      CreateTable data;
      try {
        data = this.adminClient.tables
            .create(CreateTableInput.builder().organizationName(Utils.formatEntityName(identifier.organization))
                .projectName(Utils.formatEntityName(identifier.project))
                .tableName(Utils.formatEntityName(identifier.table)).schemaKind(schemaKind).schema(schema)
                .description(description).meta(meta).useIndex(useIndex).useWarehouse(useWarehouse)
                .logRetentionSeconds(logRetention).indexRetentionSeconds(indexRetention)
                .warehouseRetentionSeconds(warehouseRetention).updateIfExists(updateIfExists).build())
            .get();
      } catch (InterruptedException | ExecutionException e) {
        throw new RuntimeException(e);
      }
      table = Table.make(this, identifier, data);
    }
    return table;
  }

  // WRITING

  /**
   * Opens the client for writes. Can be called multiple times, but make sure to
   * call ``stop`` correspondingly.
   */
  public void start() {
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
  public void stop() {
    if (this.startCount == 0) {
      throw new RuntimeException("Called stop more times than start");
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
   * server. See the BeneathClient constructor for details.
   * 
   * To enabled writes, make sure to call ``start`` on the client (and ``stop``
   * before terminating).
   * 
   * Args: instance (TableInstance): The instance to write to. You can also call
   * ``instance.write`` as a convenience wrapper. records: The records to write.
   * Can be a single record (dict) or a list of records (iterable of dict).
   */
  public void write(TableInstance instance, List<GenericRecord> records) {
    if (this.startCount == 0) {
      throw new RuntimeException("Cannot call write because the client is stopped");
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
  public void forceFlush() {
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
   */
  public Checkpointer checkpointer(String projectPath) {
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
   */
  public Checkpointer checkpointer(String projectPath, String keyPrefix, String metatableName, Boolean metatableCreate,
      String metatableDescription) {
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

  // CONSUMERS

  /**
   * Creates a consumer for the given table. Consumers make it easy to replay the
   * history of a table and/or subscribe to new changes.
   */
  public Consumer consumer(String tablePath, String subscriptionPath, Checkpointer checkpointer) {
    TableIdentifier tableIdentifier = TableIdentifier.fromPath(tablePath);
    // if (subscriptionPath == null) {}

    SubscriptionIdentifier subIdentifier = SubscriptionIdentifier.fromPath(subscriptionPath);
    if (!this.consumers.containsKey(subIdentifier)) {
      // if (checkpointer == null) {}
      Consumer consumer = new Consumer(this, tableIdentifier, subscriptionPath, checkpointer);
      consumer.init();
      this.consumers.put(subIdentifier, consumer);
    }

    return this.consumers.get(subIdentifier);
  }
}
