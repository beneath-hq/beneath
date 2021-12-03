package dev.beneath;

public class DebeziumConfig {
  public static final String DATABASE_HOSTNAME = "host.docker.internal";
  public static final String DATABASE_PORT = "5432";
  public static final String DATABASE_USER = "ericgreen";
  public static final String DATABASE_PASSWORD = "";
  public static final String DATABASE_DBNAME = "testdb";
  public static final String TABLE_INCLUDE_LIST = "public.table1, public.table2, public.table3";
  public static final String DATABASE_SERVER_NAME = "testserver";
  public static final String BENEATH_SECRET = "7HsGdXeNUygRtsZKcLGmutouBz83Fp6ksfmW3LyG2GUa";
  public static final String BENEATH_USERNAME = "ericpgreen2";
  public static final String BENEATH_PROJECT_PATH = BENEATH_USERNAME + "/debezium-postgres-" + DATABASE_DBNAME;
  public static final String BENEATH_DEBEZIUM_ROOT_TABLE_PATH = BENEATH_PROJECT_PATH + "/raw_changes";
}
