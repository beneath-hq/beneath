package dev.beneath.client;

import java.util.UUID;

public class InstanceIdAndRecordPb {
  public UUID instanceId;
  public Record record;

  InstanceIdAndRecordPb(UUID instanceId, Record record) {
    this.instanceId = instanceId;
    this.record = record;
  }
}
