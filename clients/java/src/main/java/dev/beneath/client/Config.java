package dev.beneath.client;

import java.util.Arrays;

public class Config {
  public static final Boolean DEV = Arrays.asList("dev", "development").contains(System.getenv("BENEATH_ENV"));

  public static final String JAVA_CLIENT_ID = "beneath-java";
  // TODO: Pull the version from the build.gradle file. See:
  // https://stackoverflow.com/questions/33020069/how-to-get-version-attribute-from-a-gradle-build-to-be-included-in-runtime-swing
  public static final String JAVA_CLIENT_VERSION = "1.0.1";

  public static final Integer DEFAULT_READ_BATCH_SIZE = 1000;
  public static final Integer DEFAULT_WRITE_DELAY_MS = 1000;
  public static final Integer DEFAULT_CHECKPOINT_COMMIT_DELAY_MS = 30000;
  public static final Integer MAX_RECORD_SIZE_BYTES = 8192;
  public static final Integer MAX_BATCH_SIZE_BYTES = 10000000;
  public static final Integer MAX_BATCH_SIZE_COUNT = 10000;

  public static String BENEATH_FRONTEND_HOST;
  public static String BENEATH_CONTROL_HOST;
  public static String BENEATH_GATEWAY_HOST;
  public static String BENEATH_GATEWAY_HOST_GRPC;
  static {
    if (DEV) {
      BENEATH_FRONTEND_HOST = "http://host.docker.internal:3000";
      BENEATH_CONTROL_HOST = "http://host.docker.internal:4000/graphql";
      BENEATH_GATEWAY_HOST = "http://host.docker.internal:5000";
      BENEATH_GATEWAY_HOST_GRPC = "host.docker.internal:50051";
    } else {
      BENEATH_FRONTEND_HOST = "https://beneath.dev";
      BENEATH_CONTROL_HOST = "https://control.beneath.dev";
      BENEATH_GATEWAY_HOST = "https://data.beneath.dev";
      BENEATH_GATEWAY_HOST_GRPC = "grpc.data.beneath.dev";
    }
  }
}
