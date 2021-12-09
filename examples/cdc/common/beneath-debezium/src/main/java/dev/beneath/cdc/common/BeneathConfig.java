package dev.beneath.cdc.common;

public class BeneathConfig {
  public static final String DATABASE_DBNAME = "testdb";
  public static final String BENEATH_SECRET = "7HsGdXeNUygRtsZKcLGmutouBz83Fp6ksfmW3LyG2GUa";
  public static final String BENEATH_USERNAME = "ericpgreen2";
  public static final String BENEATH_PROJECT_PATH = BENEATH_USERNAME + "/debezium-postgres-" + DATABASE_DBNAME;
  public static final String BENEATH_DEBEZIUM_ROOT_TABLE_PATH = BENEATH_PROJECT_PATH + "/raw_changes";
}
