package dev.beneath.client;

import java.util.concurrent.CompletableFuture;

import dev.beneath.client.admin.AdminClient;
import dev.beneath.client.utils.TableIdentifier;

/**
 * The main class for interacting with Beneath. Data-related features (like
 * defining tables and reading/writing data) are implemented directly on
 * `Client`, while control-plane features (like creating projects) are isolated
 * in the `admin` member.
 * 
 * Args: secret (str): A beneath secret to use for authentication. If not set,
 * uses the ``BENEATH_SECRET`` environment variable, and if that is not set
 * either, uses the secret authenticated in the CLI (stored in ``~/.beneath``).
 * dry (bool): If true, the client will not perform any mutations or writes, but
 * generally perform reads as usual. It's useful for testing.
 * 
 * The exact implication differs for different operations: Some mutations will
 * be mocked, such as creating a table, others will fail with an exception.
 * Write operations log records to the logger instead of transmitting to the
 * server. Reads generally work, but throw an exception when called on mocked
 * resources. write_delay_ms (int): The maximum amount of time to buffer written
 * records before sending a batch write request over the network. Defaults to 1
 * second (1000 ms). Writing records in batches reduces the number of requests,
 * which leads to lower cost (Beneath charges at least 1kb per request).
 */
public class Client {
  public Connection connection;
  public AdminClient adminClient;
  public Boolean dry;

  public Client(String secret, Boolean dry, Integer writeDelayMs) {
    connection = new Connection(secret);
    adminClient = new AdminClient(connection, dry);
  }

  public Table findTable(String tablePath) throws Exception {
    TableIdentifier identifier = TableIdentifier.fromPath(tablePath);
    Table table = Table.make(this, identifier, null);
    return table;
  }
}
