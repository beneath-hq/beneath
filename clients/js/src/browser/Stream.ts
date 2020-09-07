import { Cursor } from "./Cursor";
import { Connection } from "./Connection";
import { StreamQualifier, QueryLogOptions, QueryIndexOptions } from "../types";

type QueryResult<TRecord> = { cursor?: Cursor<TRecord>, error?: Error };
export type QueryLogResult<TRecord = any> = QueryResult<TRecord>;
export type QueryIndexResult<TRecord = any> = QueryResult<TRecord>;
export type WriteResult = { writeID?: string, error?: Error };

export class Stream<TRecord = any> {
  private connection: Connection;
  private streamQualifier: StreamQualifier;

  constructor(connection: Connection, streamQualifier: StreamQualifier) {
    this.connection = connection;
    this.streamQualifier = streamQualifier;
  }

  public async write(records: TRecord[]): Promise<WriteResult> {
    const { meta, error } = await this.connection.write(this.streamQualifier, records);
    if (error) {
      return { error };
    }

    return { writeID: meta?.writeID };
  }

  public async queryLog(opts?: QueryLogOptions): Promise<QueryResult<TRecord>> {
    const args = { peek: opts?.peek, limit: opts?.pageSize };
    const { meta, data, error } = await this.connection.queryLog(this.streamQualifier, args);
    if (error) {
      return { error };
    }

    let qualifier = this.streamQualifier;
    if (meta?.instanceID) {
      qualifier = { instanceID: meta.instanceID };
    }

    const cursor = new Cursor<TRecord>(this.connection, meta?.nextCursor, meta?.changeCursor, data, qualifier, opts?.pageSize);
    return { cursor };
  }

  public async queryIndex(opts?: QueryIndexOptions): Promise<QueryResult<TRecord>> {
    const args = { filter: opts?.filter, limit: opts?.pageSize };
    const { meta, data, error } = await this.connection.queryIndex(this.streamQualifier, args);
    if (error) {
      return { error };
    }

    let qualifier = this.streamQualifier;
    if (meta?.instanceID) {
      qualifier = { instanceID: meta.instanceID };
    }

    const cursor = new Cursor<TRecord>(this.connection, meta?.nextCursor, meta?.changeCursor, data, qualifier, opts?.pageSize);
    return { cursor };
  }

}
