import { SubscriptionClient } from "subscriptions-transport-ws";

import { BrowserCursor } from "./BrowserCursor";
import { BrowserConnection } from "./BrowserConnection";
import { StreamQualifier, QueryOptions, ReadOptions } from "./shared";

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

  public async query(opts?: QueryOptions): Promise<BrowserQueryResult<TRecord>> {
    // default options
    opts = opts || {};
    opts.compact = opts.compact === undefined ? true : opts.compact;

    // fetch
    const { records, error, cursor } = await this.connection.query(this.streamQualifier, {
      compact: opts.compact,
      filter: opts.filter,
      limit: opts.pageSize,
    });
    if (error) {
      return { error };
    }

    // wrap cursor and return
    const cursorWrap = new BrowserCursor<TRecord>({
      connection: this.connection,
      streamQualifier: this.streamQualifier,
      cursorType: "query",
      cursor,
      defaultPageSize: opts?.pageSize,
      initialRecords: records,
    });

    return { cursor: cursorWrap };
  }

  public async peek(opts?: ReadOptions): Promise<BrowserQueryResult<TRecord>> {
    const { records, error, cursor } = await this.connection.peek<TRecord>(this.streamQualifier, { limit: opts?.pageSize });
    if (error) {
      return { error };
    }

    // wrap cursor and return
    const cursorWrap = new BrowserCursor<TRecord>({
      connection: this.connection,
      streamQualifier: this.streamQualifier,
      cursorType: "peek",
      cursor,
      defaultPageSize: opts?.pageSize,
      initialRecords: records,
    });
    return { cursor: cursorWrap };
  }

}
