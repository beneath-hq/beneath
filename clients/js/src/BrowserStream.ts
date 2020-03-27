import { BrowserCursor } from "./BrowserCursor";
import { BrowserConnection } from "./BrowserConnection";
import { StreamQualifier, ReadOptions, QueryLogOptions, QueryIndexOptions } from "./shared";

export interface BrowserQueryResult<TRecord = any> {
  cursor?: BrowserCursor<TRecord>;
  error?: Error;
}

export class BrowserStream<TRecord = any> {
  private connection: BrowserConnection;
  private streamQualifier: StreamQualifier;

  constructor(connection: BrowserConnection, streamQualifier: StreamQualifier) {
    this.connection = connection;
    this.streamQualifier = streamQualifier;
  }

  public async queryLog(opts?: QueryLogOptions): Promise<BrowserQueryResult<TRecord>> {
    const args = { peek: opts?.peek, limit: opts?.pageSize };
    const { meta, data, error } = await this.connection.queryLog(this.streamQualifier, args);
    if (error) {
      return { error };
    }

    const cursor = new BrowserCursor<TRecord>(this.connection, this.streamQualifier, meta, data, opts?.pageSize);
    return { cursor };
  }

  public async queryIndex(opts?: QueryIndexOptions): Promise<BrowserQueryResult<TRecord>> {
    const args = { filter: opts?.filter, limit: opts?.pageSize };
    const { meta, data, error } = await this.connection.queryIndex(this.streamQualifier, args);
    if (error) {
      return { error };
    }

    const cursor = new BrowserCursor<TRecord>(this.connection, this.streamQualifier, meta, data, opts?.pageSize);
    return { cursor };
  }

}
