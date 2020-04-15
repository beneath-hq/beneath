import { BrowserStream } from "./BrowserStream";
import { BrowserConnection } from "./BrowserConnection";
import { StreamQualifier } from "./shared";

/**
 * Options passed to the `BrowserClient` constructor
 */
export interface BrowserClientOptions {
  /**
   * The secret used to authenticate to Beneath. NOTE: See
   * {@linkcode BrowserClient} for details on secrets.
   */
  secret?: string;
}

/**
 * `BrowserClient` is the root class for interfacing with Beneath from the
 * browser. It is a wrapper for the Beneath REST APIs. You can use it to read
 * from and write data to streams.
 *
 * To instantiate a new client and find a stream:
 *
 * ```js
 * let client = BrowserClient({ secret: "YOUR_SECRET" });
 * let stream = client.findStream("USERNAME/PROJECT/STREAM");
 * ```
 *
 * If your code runs in the browser (i.e. it's part of your frontend), you must
 * use a read-only secret. You can obtain a new secret from your settings page
 * on [https://beneath.dev](https://beneath.dev).
 *
 * Despite its name, `BrowserClient` can also be used outside the browser. Its
 * name derives from the fact that we will soon publish a class, `Client`, which
 * has the same interface as `BrowserClient`, but uses the non-browser
 * compatible gRPC APIs (like the Python client does) to achieve less bandwidth
 * use and higher performance.
 */
export class BrowserClient {
  public secret: string | undefined;
  private connection: BrowserConnection;

  /**
   * @param opts The connection options, see the class docs for an example
   */
  constructor(opts?: BrowserClientOptions) {
    this.secret = opts && opts.secret;
    this.connection = new BrowserConnection(this.secret);
  }

  /**
   * @param streamQualifier  Identifies the stream to find
   * @typeParam TRecord  Optional type for the records in the stream. No error
   * is thrown if it doesn't correctly correspond to the stream's schema; it is
   * for type hinting purposes only.
   */
  findStream<TRecord = any>(streamQualifier: StreamQualifier): BrowserStream<TRecord> {
    return new BrowserStream<TRecord>(this.connection, streamQualifier);
  }

}
