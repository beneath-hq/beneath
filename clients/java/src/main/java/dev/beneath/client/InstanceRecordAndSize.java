package dev.beneath.client;

import org.apache.avro.generic.GenericRecord;

public class InstanceRecordAndSize {
  public TableInstance instance;
  public GenericRecord record;
  public Integer size;

  InstanceRecordAndSize(TableInstance instance, GenericRecord record, Integer size) {
    this.instance = instance;
    this.record = record;
    this.size = size;
  }
}
