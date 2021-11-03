package dev.beneath.client;

import dev.beneath.client.utils.TableIdentifier;
import dev.beneath.type.TableSchemaKind;

/**
 * Checkpointers store (small) key-value records in a meta table (in the
 * specified project). They are useful for maintaining consumer and pipeline
 * state, such as the cursor for a subscription or the last time a scraper ran.
 * 
 * Checkpoint keys are strings and values are serialized with msgpack (supports
 * most normal Python values, but not custom classes).
 * 
 * New checkpointed values are flushed at regular intervals (every 30 seconds by
 * default). Checkpointers should not be used for storing large amounts of data.
 * Checkpointers are not currently suitable for synchronizing parallel
 * processes.
 */
public class Checkpointer {
  public TableInstance instance;
  private Client client;
  private TableIdentifier metatableIdentifer;
  private String metatableDescription;
  private Boolean create;
  private static final Integer SERVICE_CHECKPOINT_LOG_RETENTION = 60 * 60 * 6; // 6 hours
  private static final String SERVICE_CHECKPOINT_SCHEMA = """
      type Checkpoint @schema {
      key: String! @key
      value: Bytes
      }
      """;

  public Checkpointer(String metatableDescription) {
    this.metatableDescription = metatableDescription;
  }

  // CHECKPOINT TABLE

  private void stageTable() throws Exception {
    Table table;
    if (this.create || this.client.dry) {
      table = this.client.createTable(this.metatableIdentifer.toString(), SERVICE_CHECKPOINT_SCHEMA,
          this.metatableDescription, true, true, false, 0, SERVICE_CHECKPOINT_LOG_RETENTION, 0, 0,
          TableSchemaKind.GRAPHQL, "", true);
    } else {
      table = this.client.findTable(this.metatableIdentifer.toString());
    }
    if (table.primaryInstance == null) {
      throw new Exception("Expected checkpoints table to have a primary instance");
    }
    this.instance = table.primaryInstance;

    // TODO: use a logger to emit this message
    System.out.print(String.format("Using '%s' (version %i) for checkpointing", this.metatableIdentifer.toString(),
        this.instance.version));
  }
}
