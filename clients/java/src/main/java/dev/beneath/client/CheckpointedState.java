package dev.beneath.client;

import com.google.protobuf.ByteString;

public class CheckpointedState {
  private ByteString replayCursor;
  private ByteString changesCursor;

  public CheckpointedState() {
  }

  public ByteString getReplayCursor() {
    return replayCursor;
  }

  public void setReplayCursor(ByteString replayCursor) {
    this.replayCursor = replayCursor;
  }

  public ByteString getChangesCursor() {
    return changesCursor;
  }

  public void setChangesCursor(ByteString changesCursor) {
    this.changesCursor = changesCursor;
  }
}
