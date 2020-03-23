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
    const { records, error, cursor } = await this.connection.query(this.streamQualifier, opts.compact, opts.filter, opts.pageSize);
    if (error) {
      return { error };
    }

    // wrap cursor and return
    const cursorWrap = new BrowserCursor<TRecord>(this.connection, this.streamQualifier, records || [], cursor?.next, cursor?.changes);
    return { cursor: cursorWrap };
  }

  public async peek(opts?: ReadOptions): Promise<BrowserQueryResult<TRecord>> {
    opts = opts || {};
    const { records, error, cursor } = await this.connection.peek<TRecord>(this.streamQualifier, opts.pageSize);
    if (error) {
      return { error };
    }

    // wrap cursor and return
    const cursorWrap = new BrowserCursor<TRecord>(this.connection, this.streamQualifier, records || [], cursor?.next, cursor?.changes);
    return { cursor: cursorWrap };
  }

}
