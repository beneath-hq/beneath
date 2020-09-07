import { Connection, PingData, Response } from "./Connection";
import { Job } from "./Job";
import { Stream } from "./Stream";
import { DEFAULT_QUERY_WAREHOUSE_TIMEOUT_MS, DEFAULT_QUERY_WAREHOUSE_MAX_BYTES_SCANNED, JS_CLIENT_ID } from "./config";
import { StreamQualifier } from "./types";
import { PACKAGE_VERSION } from "./version";

export interface ClientOptions {
  /**
   * The secret used to authenticate to Beneath. NOTE: See
   * {@linkcode Client} for details on secrets.
   */
  secret?: string;
}

export type QueryWarehouseOptions = {
  query: string;
  dry?: boolean;
  maxBytesScanned?: number;
  timeoutMilliseconds?: number;
};

export interface QueryWarehouseResult<TRecord = any> {
  job?: Job<TRecord>;
  error?: Error;
}

/**
 * `Client` is the root class for interfacing with Beneath from the
 * browser. It is a wrapper for the Beneath REST APIs. You can use it to read
 * from and write data to streams, and to run warehouse queries.
 *
 * To instantiate a new client and find a stream:
 *
 * ```js
 * const client = Client({ secret: "YOUR_SECRET" });
 * const stream = client.findStream("USERNAME/PROJECT/STREAM");
 * ```
 *
 * If your code runs in the browser (i.e. it's part of your frontend), you must
 * use a read-only secret. You can obtain a new secret from your settings page
 * on [https://beneath.dev](https://beneath.dev).
 *
 * Note that you can use this class from outside the browser, but it uses the
 * less performant REST APIs, not the gRPC APIs (like the Python client does).
 */
export class Client {
  public secret: string | undefined;
  private connection: Connection;

  /**
   * @param opts The connection options, see the class docs for an example
   */
  constructor(opts?: ClientOptions) {
    this.secret = opts && opts.secret;
    this.connection = new Connection(this.secret);
  }

  /**
   * @param streamQualifier  Identifies the stream to find
   * @typeParam TRecord  Optional type for the records in the stream. No error
   * is thrown if it doesn't correctly correspond to the stream's schema; it is
   * for type hinting purposes only.
   */
  public findStream<TRecord = any>(streamQualifier: StreamQualifier): Stream<TRecord> {
    return new Stream<TRecord>(this.connection, streamQualifier);
  }

  /**
   * Pings the server and reports a) whether the user is authenticated, b) the library version is up-to-date
   */
  public async ping(): Promise<Response<PingData>> {
    return await this.connection.ping({ clientID: JS_CLIENT_ID, clientVersion: PACKAGE_VERSION });
  }

  /**
   * @param opts Parameters of the warehouse query to execute
   */
  public async queryWarehouse<TRecord = any>(opts: QueryWarehouseOptions): Promise<QueryWarehouseResult<TRecord>> {
    opts.maxBytesScanned = opts.maxBytesScanned ?? DEFAULT_QUERY_WAREHOUSE_MAX_BYTES_SCANNED;
    opts.timeoutMilliseconds = opts.timeoutMilliseconds ?? DEFAULT_QUERY_WAREHOUSE_TIMEOUT_MS;
    const res = await this.connection.queryWarehouse(opts);
    if (res.error || !res.data) {
      return res;
    }
    const data = res.data;
    const job = new Job<TRecord>(this.connection, data);
    return { job };
  }

}
