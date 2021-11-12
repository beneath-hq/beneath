package dev.beneath.client;

import java.util.ArrayList;
import java.util.List;

import com.google.protobuf.ByteString;

import org.apache.avro.generic.GenericRecord;

/**
 * A cursor allows you to page through the results from of a query in Beneath.
 */
public class Cursor {
  Connection connection;
  Schema schema;
  ByteString replayCursor;
  ByteString changesCursor;

  Cursor(Connection connection, Schema schema, ByteString replayCursor, ByteString changesCursor) {
    this.connection = connection;
    this.schema = schema;
    // The replay cursor, which pages through the initial query results
    this.replayCursor = replayCursor;
    // The change cursor, which can return updates since the query was started
    this.changesCursor = changesCursor;
  }

  /**
   * Returns the first record or None if the cursor is empty
   * 
   * @throws Exception
   */
  public GenericRecord readOne() throws Exception {
    List<GenericRecord> batch = this.readNext(1);
    if (batch != null) {
      for (GenericRecord record : batch) {
        return record;
      }
    }
    return null;
  }

  /**
   * Returns a new page of results and advances the replay cursor
   * 
   * @throws Exception
   */
  public List<GenericRecord> readNext(Integer limit) throws Exception {
    List<Record> batch = this.readNextReplay(limit);
    if (batch.size() == 0) {
      return null;
    }
    List<GenericRecord> records = new ArrayList<GenericRecord>();
    for (Record pb : batch) {
      records.add(this.schema.pbToRecord(pb));
    }
    return records;
  }

  private List<Record> readNextReplay(Integer limit) throws Exception {
    if (this.replayCursor == null) {
      return null;
    }
    ReadResponse response = this.connection.read(this.replayCursor, limit);
    this.replayCursor = response.getNextCursor();
    return response.getRecordsList();
  }
}
