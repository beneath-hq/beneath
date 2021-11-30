package dev.beneath.client;

public class Config {
  public static final Boolean DEV = true;

  public static final String JAVA_CLIENT_ID = "beneath-java";
  public static final String JAVA_CLIENT_VERSION = "0.0.1";

  public static final Integer DEFAULT_READ_BATCH_SIZE = 1000;
  public static final Integer DEFAULT_WRITE_DELAY_MS = 1000;
  public static final Integer DEFAULT_CHECKPOINT_COMMIT_DELAY_MS = 30000;
  public static final Integer MAX_RECORD_SIZE_BYTES = 8192;
  public static final Integer MAX_BATCH_SIZE_BYTES = 10000000;
  public static final Integer MAX_BATCH_SIZE_COUNT = 10000;

  public static final String BENEATH_FRONTEND_HOST = "http://host.docker.internal:3000";
  public static final String BENEATH_CONTROL_HOST = "http://host.docker.internal:4000/graphql";
  public static final String BENEATH_GATEWAY_HOST = "http://host.docker.internal:5000";
  public static final String BENEATH_GATEWAY_HOST_GRPC = "host.docker.internal:50051";
}
