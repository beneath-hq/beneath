import { BrowserConnection, Response } from "./BrowserConnection";
import { Record, ReadOptions, ReadResult, StreamQualifier } from "./shared";

interface BrowserCursorState<TRecord> {
  connection: BrowserConnection;
  streamQualifier: StreamQualifier;
  cursorType: "query" | "peek";
  cursor?: {
    next?: string;
    changes?: string;
  };
  defaultPageSize?: number;
  initialRecords?: Record<TRecord>[];
}

export class BrowserCursor<TRecord = any> {
  private state: BrowserCursorState<TRecord>;

  constructor(state: BrowserCursorState<TRecord>) {
    this.state = state;
  }

  public canReadNext(): boolean {
    return !!this.state.cursor?.next;
  }

  public async readNext(opts?: ReadOptions): Promise<ReadResult<TRecord>> {
    let limit = opts?.pageSize || this.state.defaultPageSize;

    // tmp contains records from initialRecords to return
    let tmp: Record<TRecord>[] | undefined;
    if (this.state.initialRecords) {
      if (!limit) {
        tmp = this.state.initialRecords;
        this.state.initialRecords = undefined;
        return { records: tmp };
      }

      if (this.state.initialRecords.length >= limit) {
        tmp = this.state.initialRecords.slice(0, limit);
        if (limit === this.state.initialRecords.length) {
          this.state.initialRecords = undefined;
        } else {
          this.state.initialRecords = this.state.initialRecords.slice(limit);
        }
        return { records: tmp };
      }

      tmp = this.state.initialRecords;
      this.state.initialRecords = undefined;
      limit -= tmp.length;
    }

    // check can fetch more
    if (!this.state.cursor?.next) {
      if (tmp) {
        return {records: tmp};
      }
      return { error: Error("reached end of cursor") };
    }

    // fetch more
    let resp: Response<TRecord>;
    if (this.state.cursorType === "query") {
      resp = await this.state.connection.query<TRecord>(this.state.streamQualifier, { limit, cursor: this.state.cursor.next });
    } else if (this.state.cursorType === "peek") {
      resp = await this.state.connection.peek<TRecord>(this.state.streamQualifier, { limit, cursor: this.state.cursor.next });
    } else {
      throw Error(`unexpected value for this.state.cursor.type: <${this.state.cursorType}>`);
    }

    // parse resp
    if (resp.error) {
      return { error: resp.error };
    }

    if (!resp.cursor) {
      this.state.cursor.next = undefined;
    } else {
      this.state.cursor.next = resp.cursor.next;
    }

    return { records: resp.records };
  }

}
