package dev.beneath.cdc.postgres;

public class CdcConfig {
  public static final String DATABASE_HOSTNAME = System.getenv("DATABASE_HOSTNAME");
  public static final String DATABASE_PORT = System.getenv("DATABASE_PORT");
  public static final String DATABASE_USER = System.getenv("DATABASE_USER");
  public static final String DATABASE_PASSWORD = System.getenv("DATABASE_PASSWORD");
  public static final String DATABASE_DBNAME = System.getenv("DATABASE_DBNAME");
  public static final String TABLE_INCLUDE_LIST = System.getenv("TABLE_INCLUDE_LIST");
  public static final String DATABASE_SERVER_NAME = System.getenv("DATABASE_SERVER_NAME");
  public static final String BENEATH_SECRET = System.getenv("BENEATH_SECRET");
  public static final String BENEATH_USERNAME = System.getenv("BENEATH_USERNAME");
  public static final String BENEATH_PROJECT_PATH = BENEATH_USERNAME + "/cdc-postgres-" + DATABASE_DBNAME;
  public static final String BENEATH_DEBEZIUM_ROOT_TABLE_PATH = BENEATH_PROJECT_PATH + "/raw_changes";
  public static final Integer DEFAULT_WRITE_DELAY_MS = 1000;
}
