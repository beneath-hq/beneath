package dev.beneath;

import java.util.function.Consumer;

import io.debezium.relational.history.AbstractDatabaseHistory;
import io.debezium.relational.history.DatabaseHistoryException;
import io.debezium.relational.history.HistoryRecord;

// A required dependency for the MySQL and SQL Server Debezium connectors.
// See all official implementations:
// https://github.com/debezium/debezium/tree/main/debezium-core/src/main/java/io/debezium/relational/history
// See Pulsar implementation:
// https://github.com/apache/pulsar/blob/master/pulsar-io/debezium/core/src/main/java/org/apache/pulsar/io/debezium/PulsarDatabaseHistory.java
// https://github.com/apache/pulsar/blob/master/pulsar-io/debezium/core/src/test/java/org/apache/pulsar/io/debezium/PulsarDatabaseHistoryTest.java
public class BeneathDatabaseHistory extends AbstractDatabaseHistory {

  @Override
  public boolean exists() {
    // TODO Auto-generated method stub
    return false;
  }

  @Override
  public boolean storageExists() {
    // TODO Auto-generated method stub
    return false;
  }

  @Override
  protected void storeRecord(HistoryRecord record) throws DatabaseHistoryException {
    // TODO Auto-generated method stub

  }

  @Override
  protected void recoverRecords(Consumer<HistoryRecord> records) {
    // TODO Auto-generated method stub

  }
}
