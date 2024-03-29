package dev.beneath.client.utils;

import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutionException;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * AIODelayBuffer provides a means of buffering and flushing data based on the
 * size of buffered values or time passed since the first value was written to
 * the buffer.
 * 
 * We use it to buffer writes for `maxDelayMs` before sending them in a single
 * batched request over the network (with forced buffer flushes at
 * `maxBufferSize`).
 * 
 * It only lets one buffer be open at any moment. If a write is attempted when
 * the buffer is full or flushing, the write will not return until the flush of
 * the previous buffer has completed.
 * 
 * This class is NOT thread-safe.
 */
public abstract class AIODelayBuffer<T> {
  public Integer maxDelay;
  protected Integer maxRecordSize;
  protected Integer maxBufferSize;
  protected Integer maxBufferCount;

  private CompletableFuture<Void> delayTask;
  private CompletableFuture<Void> delayedFlushTask;

  public Boolean running;
  private Boolean flushing;
  private Integer bufferSize;
  private Integer bufferCount;

  private static final Logger LOGGER = LoggerFactory.getLogger(AIODelayBuffer.class);

  protected AIODelayBuffer(Integer maxDelayMs, Integer maxRecordSize, Integer maxBufferSize, Integer maxBufferCount) {
    this.maxDelay = maxDelayMs / 1000;
    this.maxRecordSize = maxRecordSize;
    this.maxBufferSize = maxBufferSize;
    this.maxBufferCount = maxBufferCount;

    this.running = false;
    this.flushing = false;
    this.bufferSize = 0;
    this.bufferCount = 0;
    this.reset();
  }

  protected abstract void reset();

  protected abstract void merge(T value);

  protected abstract void flush();

  public void start() {
    if (this.running) {
      throw new RuntimeException("Already called start");
    }
    this.running = true;
  }

  public void stop() {
    this.running = false;
    this.forceFlush();
  }

  /**
   * Adds value to the buffer. When an awaited call to write returns, the value
   * has been added to the buffer, but not been flushed yet. If you wish to wait
   * until the write has been flushed, await the task returned by write. For
   * example:
   * 
   * task = buffer.write(value=..., size=..).get(); task.get()
   */
  public CompletableFuture<Void> write(T value, Integer size) {
    // check open
    if (this.running == false) {
      throw new RuntimeException("Cannot call write because the buffer is closed");
    }

    // check value is within acceptable record size
    if (size > this.maxRecordSize) {
      throw new RuntimeException(
          String.format("Value exceeds maximum record size (size=%d, max_record_size=%d, value=%s)", size,
              this.maxRecordSize, value));
    }

    // trigger/wait for flush if a) a flush is in progress, or b) value would cause
    // size overflow
    Integer loops = 0;
    while (this.flushing || (this.bufferSize + size > this.maxBufferSize)
        || (this.bufferCount == this.maxBufferCount)) {
      assert this.delayedFlushTask != null;
      this.forceFlush();
      loops += 1;
      if (loops > 5) {
        LOGGER.warn(String
            .format("Unfortunate scheduling blocked write to buffer %d times (try to limit concurrent writes)", loops));
      }
    }

    // now we know we're not flushing and the value fits;
    // and execution will not be "interrupted" until next "await"

    // add to buffer
    this.merge(value);
    this.bufferSize += size;
    this.bufferCount += 1;

    // if a delayed flush isn't already scheduled for this batch, schedule it now
    if (this.delayedFlushTask == null) {
      this.delayTask = CompletableFuture.runAsync(() -> {
        try {
          Thread.sleep(maxDelay);
        } catch (Exception e) {
          LOGGER.error("Interrupted the task.");
        }
      });
      this.delayedFlushTask = CompletableFuture.runAsync(() -> this.delayedFlush()).exceptionally(e -> {
        LOGGER.error("Error in buffer flush background loop: %s", e.getMessage());
        return null;
      });
    }

    return this.delayedFlushTask;
  }

  public void forceFlush() {
    // cancel the delay
    if (this.delayTask != null) {
      this.delayTask.cancel(true);
    }
    // proceed with flush
    if (this.delayedFlushTask != null) {
      try {
        this.delayedFlushTask.get();
      } catch (InterruptedException | ExecutionException e) {
        throw new RuntimeException(e);
      }
    }
  }

  private void delayedFlush() {
    // wait for delay
    if (this.delayTask != null) {
      try {
        delayTask.get();
      } catch (Exception e) {
      }
    }
    // flush
    this.flushing = true;
    this.flush();
    this.reset();
    this.flushing = false;
    this.bufferSize = 0;
    this.bufferCount = 0;
    this.delayTask = null;
    this.delayedFlushTask = null;
  }
}
