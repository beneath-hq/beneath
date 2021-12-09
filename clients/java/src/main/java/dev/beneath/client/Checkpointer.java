package dev.beneath.client;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Map.Entry;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

import org.apache.avro.Schema;
import org.apache.avro.generic.GenericRecord;
import org.apache.avro.generic.GenericRecordBuilder;
import org.msgpack.jackson.dataformat.MessagePackFactory;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import dev.beneath.client.utils.AIODelayBuffer;
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
  private BeneathClient client;
  private TableIdentifier metatableIdentifier;
  private String metatableDescription;
  private Boolean create;
  private Writer writer;
  private Map<String, Object> cache;
  private ObjectMapper objectMapper;
  private static final Integer SERVICE_CHECKPOINT_LOG_RETENTION = 60 * 60 * 6; // 6 hours
  private static final String SERVICE_CHECKPOINT_SCHEMA = """
      type Checkpoint @schema {
      key: String! @key
      value: Bytes
      }
      """;
  private static final Logger LOGGER = LoggerFactory.getLogger(Checkpointer.class);

  public Checkpointer(BeneathClient client, TableIdentifier metatableIdentifier, Boolean metatableCreate,
      String metatableDescription) {
    this.client = client;
    this.metatableIdentifier = metatableIdentifier;
    this.metatableDescription = metatableDescription;
    this.create = metatableCreate;
    this.writer = new Writer(this);
    this.cache = new HashMap<String, Object>();
    this.objectMapper = new ObjectMapper(new MessagePackFactory());
  }

  /**
   * Gets a checkpointed value
   */
  public Object get(String key) {
    return this.get(key, null);
  }

  /**
   * Gets a checkpointed value
   */
  public Object get(String key, Object defaultValue) {
    if (this.cache.containsKey(key)) {
      return this.cache.get(key);
    }

    if (this.client.dry) {
      return defaultValue;
    }

    String filter = String.format("{\"key\": \"%s\"}", key);
    Cursor cursor = this.instance.queryIndex(filter);

    Object value = defaultValue;
    GenericRecord record = cursor.readOne();
    if (record != null) {
      byte[] bytes = ((ByteBuffer) record.get("value")).array();
      try {
        value = objectMapper.readValue(bytes, Object.class);
      } catch (IOException e) {
        throw new RuntimeException(e);
      }
    }

    // checking self._cache again because of awaits (and we'd rather serve a recent
    // local set)
    if (!this.cache.containsKey(key)) {
      this.cache.put(key, value);
    }

    return this.cache.get(key);
  }

  /**
   * Sets a checkpoint value. Value will be encoded with msgpack.
   */
  public void set(String key, Object value) {
    if (!this.writer.running) {
      throw new RuntimeException("Cannot call 'set' on checkpointer because the client is stopped");
    }
    this.cache.put(key, value);
    this.writer.write(key, value);
  }

  // START/STOP (called by client)

  public void start() {
    if (this.instance == null) {
      this.stageTable();
    }
    this.writer.start();
  }

  public void stop() {
    this.writer.stop();
  }

  // CHECKPOINT TABLE

  private void stageTable() {
    Table table;
    if (this.create || this.client.dry) {
      table = this.client.createTable(this.metatableIdentifier.toString(), SERVICE_CHECKPOINT_SCHEMA,
          this.metatableDescription, true, true, false, 0, SERVICE_CHECKPOINT_LOG_RETENTION, 0, null,
          TableSchemaKind.GRAPHQL, "", true);
    } else {
      table = this.client.findTable(this.metatableIdentifier.toString());
    }
    if (table.primaryInstance == null) {
      throw new RuntimeException("Expected checkpoints table to have a primary instance");
    }
    this.instance = table.primaryInstance;

    LOGGER.info("Using '{}' (version {}) for checkpointing", this.metatableIdentifier.toString(),
        this.instance.version);
  }

  // CHECKPOINT WRITER

  // probably should make these private and add getters/setters
  class CheckpointKeyValue {
    public String key;
    public Object checkpoint;

    CheckpointKeyValue(String key, Object checkpoint) {
      this.key = key;
      this.checkpoint = checkpoint;
    }
  }

  class Writer extends AIODelayBuffer<CheckpointKeyValue> {
    Checkpointer checkpointer;
    Map<String, Object> checkpoints; // this is the buffer

    Writer(Checkpointer checkpointer) {
      super(Config.DEFAULT_CHECKPOINT_COMMIT_DELAY_MS, Integer.MAX_VALUE, Integer.MAX_VALUE, Integer.MAX_VALUE);
      this.checkpointer = checkpointer;
      this.checkpoints = new HashMap<String, Object>();
    }

    @Override
    protected void reset() {
      if (this.checkpoints != null) {
        this.checkpoints.clear();
      }
    }

    @Override
    protected void merge(CheckpointKeyValue value) {
      this.checkpoints.put(value.key, value.checkpoint);
    }

    @Override
    protected void flush() {
      // iterate through map and serialize w/ messagepack
      List<GenericRecord> records = new ArrayList<GenericRecord>();
      Schema schema = new Schema.Parser().parse(instance.table.schema.parsedAvro.toString());
      for (Entry<String, Object> checkpoint : checkpoints.entrySet()) {
        try {
          ByteBuffer bytes;
          bytes = ByteBuffer.wrap(objectMapper.writeValueAsBytes(checkpoint.getValue()));
          GenericRecord genericRecord = new GenericRecordBuilder(schema).set("key", checkpoint.getKey())
              .set("value", bytes).build();
          records.add(genericRecord);
        } catch (JsonProcessingException e) {
          e.printStackTrace();
        }
      }

      // issue network request
      if (this.checkpointer.instance != null) {
        try {
          this.checkpointer.client.write(this.checkpointer.instance, records);
        } catch (Exception e) {
          e.printStackTrace();
        }
      }
    }

    public void write(String key, Object checkpoint) {
      CheckpointKeyValue checkpointKeyValue = new CheckpointKeyValue(key, checkpoint);
      super.write(checkpointKeyValue, 0);
    }
  }
}
