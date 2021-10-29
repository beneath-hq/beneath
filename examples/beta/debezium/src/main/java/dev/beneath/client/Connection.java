package dev.beneath.client;

import com.apollographql.apollo.ApolloClient;

import io.grpc.Channel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.Metadata;

/**
 * Encapsulates network connectivity to Beneath
 */
public class Connection {
  private String secret;
  private Boolean connected;
  private Metadata requestMetadata;
  private Channel channel;
  private GatewayGrpc.GatewayBlockingStub blockingStub;
  private GatewayGrpc.GatewayStub asyncStub;
  private PingResponse pong;
  public ApolloClient apolloClient;

  // TODO: move constants to a config file
  private static final Boolean CONFIG_DEVELOPMENT = true;
  private static final String JAVA_CLIENT_ID = "beneath-java";
  private static final String JAVA_CLIENT_VERSION = "0.0.1";
  private static final String BENEATH_CONTROL_HOST = "http://host.docker.internal:4000/graphql";
  // private String BENEATH_GATEWAY_HOST = "http://host.docker.internal:5000";
  // private String BENEATH_GATEWAY_HOST_GRPC = "host.docker.internal:50051";

  public Connection(String secret) {
    this.secret = secret;
    this.connected = false;

    // construct metadata
    this.requestMetadata = new Metadata();
    this.requestMetadata.put(Metadata.Key.of("authorization", Metadata.ASCII_STRING_MARSHALLER),
        String.format("Bearer %s", this.secret));
  }

  // GRPC CONNECTIVITY

  // TODO: make this an async function
  public void ensureConnected(Boolean check_authenticated) throws Exception {
    if (!connected) {
      createGrpcConnection("host.docker.internal", 50051);
      PingResponse pong = ping();
      checkPongStatus(pong);
      this.pong = pong;
      connected = true;
    }
    if (check_authenticated) {
      if (!this.pong.getAuthenticated()) {
        throw new AuthenticationException("You must authenticate with 'beneath auth' or by setting BENEATH_SECRET");
      }
    }
  }

  private void createGrpcConnection(String host, Integer port) {
    Boolean insecure = true;
    if (insecure) {
      this.channel = ManagedChannelBuilder.forAddress(host, port).usePlaintext().build();
    } else {
      // TODO: create a secure channel
    }

    // TODO: pass in requestMetadata (look at a "ClientInterceptor")
    this.blockingStub = GatewayGrpc.newBlockingStub(this.channel).withCompression("gzip");
    this.asyncStub = GatewayGrpc.newStub(this.channel).withCompression("gzip");

    this.connected = true;
  }

  private static void checkPongStatus(PingResponse pong) throws Exception {
    if (CONFIG_DEVELOPMENT) {
      return;
    }
    if (pong.getVersionStatus() == "warning") {
      // TODO: emit warning
    } else if (pong.getVersionStatus() == "deprecated") {
      throw new Exception(
          String.format("This version (%s) of the Beneath java library is out-of-date (recommended: %s).",
              JAVA_CLIENT_VERSION, pong.getRecommendedVersion()));
    }
  }

  // TODO: make this an async function
  private PingResponse ping() {
    PingRequest request = PingRequest.newBuilder().setClientId(JAVA_CLIENT_ID).setClientVersion(JAVA_CLIENT_VERSION)
        .build();

    PingResponse response = this.blockingStub.ping(request);

    return response;
  }

  // CONTROL PLANE

  public void createGraphQlConnection() {
    this.apolloClient = ApolloClient.builder().serverUrl(BENEATH_CONTROL_HOST).build();
  }

}