package dev.beneath.client;

import org.apache.avro.generic.GenericRecord;

import java.util.ArrayList;
import java.util.List;
import java.util.Map.Entry;

import dev.beneath.client.utils.AIODelayBuffer;

/**
 * Override of AIODelayBuffer designed to buffer and write to multiple instances
 * at once
 */
public class DryWriter extends AIODelayBuffer<InstanceRecordAndSize> {
  private Client client;
  private List<InstanceRecordAndSize> records;
  private Integer total;

  protected DryWriter(Client client, Integer maxDelayMs) {
    super(maxDelayMs, Config.MAX_RECORD_SIZE_BYTES, Config.MAX_BATCH_SIZE_BYTES, Config.MAX_BATCH_SIZE_COUNT);
    this.client = client;
    this.total = 0;
    this.records = new ArrayList<InstanceRecordAndSize>();
  }

  @Override
  protected void reset() {
    if (records != null) {
      this.records.clear();
    }
  }

  @Override
  protected void merge(InstanceRecordAndSize value) {
    this.records.add(value);
  }

  @Override
  protected void flush() {
    // TODO: use a real logger
    System.out.println(String.format("Flushing %d buffered records", this.records.size()));
    for (InstanceRecordAndSize r : records) {
      System.out.println(String.format("Flushed record (table=%s, size=%d bytes): %s",
          r.instance.table.identifier.toString(), r.size, r.record));
      this.total += 1;
    }
    System.out.println(String.format("Flushed %d records (%d total during session)", this.records.size(), this.total));
  }

  public void write(TableInstance instance, List<GenericRecord> records) throws Exception {
    for (GenericRecord record : records) {
      Entry<Record, Integer> tuple = instance.table.schema.recordToPb(record);
      InstanceRecordAndSize value = new InstanceRecordAndSize(instance, record, tuple.getValue());
      super.write(value, tuple.getValue()).get();
    }
  }
}
