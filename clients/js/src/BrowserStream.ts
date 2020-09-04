import { BrowserCursor } from "./BrowserCursor";
import { BrowserConnection, WriteMeta } from "./BrowserConnection";
import { StreamQualifier, QueryLogOptions, QueryIndexOptions } from "./shared";

export interface BrowserWriteResult {
  writeID?: string;
  error?: Error;
}

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

  public async write(records: TRecord[]): Promise<BrowserWriteResult> {
    const { meta, error } = await this.connection.write(this.streamQualifier, records);
    if (error) {
      return { error };
    }

    return { writeID: meta?.writeID };
  }

  public async queryLog(opts?: QueryLogOptions): Promise<BrowserQueryResult<TRecord>> {
    const args = { peek: opts?.peek, limit: opts?.pageSize };
    const { meta, data, error } = await this.connection.queryLog(this.streamQualifier, args);
    if (error) {
      return { error };
    }

    let qualifier = this.streamQualifier;
    if (meta?.instanceID) {
      qualifier = { instanceID: meta.instanceID };
    }

    const cursor = new BrowserCursor<TRecord>(this.connection, meta?.nextCursor, meta?.changeCursor, data, qualifier, opts?.pageSize);
    return { cursor };
  }

  public async queryIndex(opts?: QueryIndexOptions): Promise<BrowserQueryResult<TRecord>> {
    const args = { filter: opts?.filter, limit: opts?.pageSize };
    const { meta, data, error } = await this.connection.queryIndex(this.streamQualifier, args);
    if (error) {
      return { error };
    }

    let qualifier = this.streamQualifier;
    if (meta?.instanceID) {
      qualifier = { instanceID: meta.instanceID };
    }

    const cursor = new BrowserCursor<TRecord>(this.connection, meta?.nextCursor, meta?.changeCursor, data, qualifier, opts?.pageSize);
    return { cursor };
  }

}
