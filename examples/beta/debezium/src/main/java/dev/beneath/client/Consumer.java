package dev.beneath.client;

import dev.beneath.client.utils.JsonUtils;
import dev.beneath.client.utils.TableIdentifier;

/**
 * Consumers are used to replay/subscribe to a table. If the consumer is
 * initialized with a project and subscription name, it will checkpoint its
 * progress to avoid reprocessing the same data every time the process starts.
 */
public class Consumer {
  /**
   * The table instance the consumer is subscribed to
   */
  public TableInstance instance;

  /**
   * The cursor used to replay and subscribe the table. You can use it to get the
   * current state of the the underlying replay and changes cursors.
   */
  public Cursor cursor;

  private BeneathClient client;
  private TableIdentifier tableIdentifier;
  private Integer version;
  // private Integer batchSize;
  private Checkpointer checkpointer;
  private String subscriptionName;

  public Consumer(BeneathClient client, TableIdentifier tableIdentifier, String subscriptionName,
      Checkpointer checkpointer) {
    this.client = client;
    this.tableIdentifier = tableIdentifier;
    this.subscriptionName = subscriptionName;
    // this.batchSize = Config.DEFAULT_READ_BATCH_SIZE;
  }

  public void init() {
    Table table = this.client.findTable(tableIdentifier.toString());
    if (this.version != null) {
      // this.instance = table.findInstance(version);
    } else {
      this.instance = table.primaryInstance;
      if (this.instance == null) {
        throw new RuntimeException(String.format("Cannot consume table %s because it doesn't have a primary instance",
            this.tableIdentifier.toString()));
      }
    }
    this.initCursor();
  }

  // CURSORS / CHECKPOINTS

  private String makeSubscriptionCursorKey() {
    return this.subscriptionName + ":" + this.instance.instanceId.toString() + ":cursor";
  }

  private void initCursor() {
    Boolean reset = false;
    if (!reset && this.checkpointer != null) {
      String checkpointedStateAsString = (String) this.checkpointer.get(makeSubscriptionCursorKey());
      if (checkpointedStateAsString != null) {
        CheckpointedState state = JsonUtils.deserialize(checkpointedStateAsString, CheckpointedState.class);
        this.cursor = this.instance.table.restoreCursor(state.getReplayCursor(), state.getChangesCursor());
      }
      return;
    }

    this.cursor = this.instance.queryLog();
    if (reset) {
      this.checkpoint();
    }
  }

  private void checkpoint() {
    if (this.checkpointer == null) {
      return;
    }
    CheckpointedState checkpointedState = new CheckpointedState();
    if (this.cursor.replayCursor != null) {
      checkpointedState.setReplayCursor(this.cursor.replayCursor);
    }
    if (this.cursor.changesCursor != null) {
      checkpointedState.setChangesCursor(this.cursor.changesCursor);
    }
    String checkpointedStateAsString = JsonUtils.serialize(checkpointedState);
    this.checkpointer.set(makeSubscriptionCursorKey(), checkpointedStateAsString);
  }
}
