package dev.beneath.client;

import java.util.Collection;
import java.util.List;
import java.util.Map.Entry;
import java.util.UUID;

import com.google.common.collect.ArrayListMultimap;
import com.google.common.collect.Multimap;
import com.google.protobuf.ByteString;

import org.apache.avro.generic.GenericRecord;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import dev.beneath.client.utils.AIODelayBuffer;
import dev.beneath.client.utils.Utils;

/**
 * Override of AIODelayBuffer designed to buffer and write to multiple instances
 * at once
 */
public class Writer extends AIODelayBuffer<InstanceIdAndRecordPb> {
  private Client client;
  private Multimap<UUID, Record> records;
  private Integer total;

  private static final Logger LOGGER = LoggerFactory.getLogger(Writer.class);

  protected Writer(Client client, Integer maxDelayMs) {
    super(maxDelayMs, Config.MAX_BATCH_SIZE_BYTES, Config.MAX_BATCH_SIZE_BYTES, Config.MAX_BATCH_SIZE_COUNT);
    this.client = client;
    this.records = ArrayListMultimap.create();
    this.total = 0;
  }

  @Override
  protected void reset() {
    if (records != null) {
      this.records.clear();
    }
  }

  @Override
  protected void merge(InstanceIdAndRecordPb value) {
    this.records.put(value.instanceId, value.record);
  }

  @Override
  protected void flush() {
    Integer count = 0;
    for (UUID key : this.records.keySet()) {
      ByteString instanceId = ByteString.copyFrom(Utils.uuidToBytes(key));
      Collection<Record> values = this.records.get(key);
      InstanceRecords instanceRecords = InstanceRecords.newBuilder().setInstanceId(instanceId).addAllRecords(values)
          .build();
      try {
        this.client.connection.write(instanceRecords);
      } catch (Exception e) {
        e.printStackTrace();
      }
      count += values.size();
    }
    this.total += count;
    LOGGER.info("Flushed {} records to {} instances ({} total during session)", count, this.records.size(), this.total);
  }

  public void write(TableInstance instance, List<GenericRecord> records) throws Exception {
    for (GenericRecord record : records) {
      Entry<Record, Integer> tuple = instance.table.schema.recordToPb(record);
      InstanceIdAndRecordPb value = new InstanceIdAndRecordPb(instance.instanceId, tuple.getKey());
      super.write(value, tuple.getValue());
    }
  }
}
