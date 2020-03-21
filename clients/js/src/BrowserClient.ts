import { BrowserStream } from "./BrowserStream";
import { BrowserConnection } from "./BrowserConnection";
import { StreamQualifier } from "./shared";

export interface BrowserClientOptions {
  secret?: string;
}

export class BrowserClient {
  public secret: string | undefined;
  private connection: BrowserConnection;

  constructor(opts?: BrowserClientOptions) {
    this.secret = opts && opts.secret;
    this.connection = new BrowserConnection(this.secret);
  }

  findStream<TRecord = any>(streamQualifier: StreamQualifier) {
    return new BrowserStream<TRecord>(this.connection, streamQualifier);
  }

}
