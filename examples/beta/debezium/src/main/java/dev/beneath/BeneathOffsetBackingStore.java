package dev.beneath;

import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.util.Collection;
import java.util.HashMap;
import java.util.Map;
import java.util.Map.Entry;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.Future;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import org.apache.kafka.connect.runtime.WorkerConfig;
import org.apache.kafka.connect.storage.OffsetBackingStore;
import org.apache.kafka.connect.util.Callback;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import dev.beneath.client.BeneathClient;
import dev.beneath.client.Checkpointer;
import dev.beneath.client.Config;

public class BeneathOffsetBackingStore implements OffsetBackingStore {
  private static final Logger LOGGER = LoggerFactory.getLogger(BeneathOffsetBackingStore.class);
  private BeneathClient client;
  private Checkpointer checkpointer;

  // TODO: preferably, make the BeneathClient a parameter in the constructor
  public BeneathOffsetBackingStore() {
    try {
      this.client = new BeneathClient(DebeziumConfig.BENEATH_SECRET, false, Config.DEFAULT_WRITE_DELAY_MS);
    } catch (Exception e) {
    }
  }

  /**
   * Start this offset store.
   */
  @Override
  public void start() {
    LOGGER.info("Starting BeneathOffsetBackingStore");
    try {
      this.checkpointer = client.checkpointer(DebeziumConfig.BENEATH_PROJECT_PATH);
      client.start();
    } catch (Exception e) {
      throw new RuntimeException("Failed to setup the Beneath checkpointer", e);
    }
    LOGGER.info("Finished reading offsets topic and starting BeneathOffsetBackingStore");
  }

  /**
   * Stop the backing store. Implementations should attempt to shutdown
   * gracefully, but not block indefinitely.
   */
  @Override
  public void stop() {
    LOGGER.info("Stopping BeneathOffsetBackingStore");
    try {
      // We only flush the client; we don't stop it. Because this is just the offset
      // store and we'll want to keep the client up for non-offset activity.
      client.forceFlush();
    } catch (Exception e) {
      throw new RuntimeException("Failed to flush offsets to Beneath");
    }
    LOGGER.info("Stopped BeneathOffsetBackingStore");
  }

  /**
   * Get the values for the specified keys
   * 
   * @param keys list of keys to look up
   * @return future for the resulting map from key to value
   */
  @Override
  public Future<Map<ByteBuffer, ByteBuffer>> get(Collection<ByteBuffer> keys) {
    CompletableFuture<Map<ByteBuffer, ByteBuffer>> future = new CompletableFuture<Map<ByteBuffer, ByteBuffer>>();
    Map<ByteBuffer, ByteBuffer> offsets = new HashMap<>();
    for (ByteBuffer key : keys) {
      try {
        String keyString = byteBufferToString(key);
        keyString = convertJsonKeyToCustomKey(keyString);
        byte[] valueBytes = (byte[]) this.checkpointer.get(keyString);
        ByteBuffer value = valueBytes != null ? ByteBuffer.wrap(valueBytes) : null;
        offsets.put(key, value);
      } catch (Exception e) {
        throw new RuntimeException(e);
      }
    }
    future.complete(offsets);
    return future;
  }

  /**
   * Set the specified keys and values.
   * 
   * @param values   map from key to value
   * @param callback callback to invoke on completion
   * @return void future for the operation
   */
  @Override
  public Future<Void> set(Map<ByteBuffer, ByteBuffer> values, Callback<Void> callback) {
    for (Entry<ByteBuffer, ByteBuffer> offset : values.entrySet()) {
      try {
        // Beneath requires that checkpoint keys are strings.
        String keyString = byteBufferToString(offset.getKey());
        // By default, Debezium offset keys are JSON. Because Beneath doesn't currently
        // support JSON keys, we convert the key to a Beneath-friendly format.
        keyString = convertJsonKeyToCustomKey(keyString);
        byte[] value = offset.getValue().array();
        this.checkpointer.set(keyString, value);
      } catch (Exception e) {
        throw new RuntimeException(e);
      }
    }

    // TODO: what's the point of the return future? should I use the callback
    // parameter for anything?
    return new CompletableFuture<Void>();
  }

  /**
   * Configure class with the given key-value pairs
   * 
   * @param config can be DistributedConfig or StandaloneConfig
   */
  @Override
  public void configure(WorkerConfig config) {
    // ...
  }

  private String byteBufferToString(ByteBuffer byteBuffer) {
    return new String(byteBuffer.array(), StandardCharsets.UTF_8);
  }

  private String convertJsonKeyToCustomKey(String keyString) throws Exception {
    JsonNode jsonNode = new ObjectMapper().readTree(keyString);
    String name = jsonNode.get("payload").get(0).asText();
    String databaseServerName = jsonNode.get("payload").get(1).get("server").asText();
    return String.format("%s:%s:offset", name, databaseServerName);
  }
}
