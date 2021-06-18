import { Cursor } from "./Cursor";
import { Connection } from "./Connection";
import { TableQualifier } from "./types";

export type QueryIndexOptions = { filter?: string, pageSize?: number };
export type QueryLogOptions = { peek?: boolean, pageSize?: number };

type QueryResult<TRecord> = { cursor?: Cursor<TRecord>, error?: Error };
export type QueryLogResult<TRecord = any> = QueryResult<TRecord>;
export type QueryIndexResult<TRecord = any> = QueryResult<TRecord>;
export type WriteResult = { writeID?: string, error?: Error };

export class Table<TRecord = any> {
  private connection: Connection;
  private tableQualifier: TableQualifier;

  constructor(connection: Connection, tableQualifier: TableQualifier) {
    this.connection = connection;
    this.tableQualifier = tableQualifier;
  }

  public async write(records: TRecord[]): Promise<WriteResult> {
    const { meta, error } = await this.connection.write(this.tableQualifier, records);
    if (error) {
      return { error };
    }

    return { writeID: meta?.writeID };
  }

  public async queryLog(opts?: QueryLogOptions): Promise<QueryLogResult<TRecord>> {
    const args = { peek: opts?.peek, limit: opts?.pageSize };
    const { meta, data, error } = await this.connection.queryLog(this.tableQualifier, args);
    if (error) {
      return { error };
    }

    let qualifier = this.tableQualifier;
    if (meta?.instanceID) {
      qualifier = { instanceID: meta.instanceID };
    }

    const cursor = new Cursor<TRecord>(this.connection, meta?.nextCursor, meta?.changeCursor, data, qualifier, opts?.pageSize);
    return { cursor };
  }

  public async queryIndex(opts?: QueryIndexOptions): Promise<QueryIndexResult<TRecord>> {
    const args = { filter: opts?.filter, limit: opts?.pageSize };
    const { meta, data, error } = await this.connection.queryIndex(this.tableQualifier, args);
    if (error) {
      return { error };
    }

    let qualifier = this.tableQualifier;
    if (meta?.instanceID) {
      qualifier = { instanceID: meta.instanceID };
    }

    const cursor = new Cursor<TRecord>(this.connection, meta?.nextCursor, meta?.changeCursor, data, qualifier, opts?.pageSize);
    return { cursor };
  }

}
