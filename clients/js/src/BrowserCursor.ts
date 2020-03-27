import { BrowserConnection, Response, ResponseMeta } from "./BrowserConnection";
import { Record, ReadOptions, ReadResult, StreamQualifier, SubscribeOptions } from "./shared";

export class BrowserCursor<TRecord = any> {
  private connection: BrowserConnection;
  private streamQualifier: StreamQualifier;
  private nextCursor?: string;
  private changeCursor?: string;
  private defaultPageSize?: number;
  private initialData?: Record<TRecord>[];

  constructor(connection: BrowserConnection, streamQualifier: StreamQualifier, meta?: ResponseMeta, data?: Record<TRecord>[], defaultPageSize?: number) {
    this.connection = connection;
    this.streamQualifier = streamQualifier;
    this.nextCursor = meta?.nextCursor;
    this.changeCursor = meta?.changeCursor;
    this.defaultPageSize = defaultPageSize;

    if (meta?.instanceID) {
      this.streamQualifier = { instanceID: meta.instanceID };
    }

    this.initialData = data;
  }

  public hasNext(): boolean {
    return !!this.nextCursor;
  }

  public hasNextChanges(): boolean {
    return !!this.changeCursor;
  }

  public async readNext(opts?: ReadOptions): Promise<ReadResult<TRecord>> {
    let limit = opts?.pageSize || this.defaultPageSize;

    // tmp contains records from initialData to return
    let tmp: Record<TRecord>[] | undefined;
    if (this.initialData) {
      if (!limit) {
        tmp = this.initialData;
        this.initialData = undefined;
        return { data: tmp };
      }

      if (this.initialData.length >= limit) {
        tmp = this.initialData.slice(0, limit);
        if (limit === this.initialData.length) {
          this.initialData = undefined;
        } else {
          this.initialData = this.initialData.slice(limit);
        }
        return { data: tmp };
      }

      tmp = this.initialData;
      this.initialData = undefined;
      limit -= tmp.length;
    }

    // check can fetch more
    if (!this.nextCursor) {
      if (tmp) {
        return { data: tmp };
      }
      return { error: Error("reached end of cursor") };
    }

    // fetch more
    const { meta, data, error } = await this.connection.read<TRecord>(this.streamQualifier, { limit, cursor: this.nextCursor });
    if (error) {
      return { error };
    }

    this.nextCursor = meta?.nextCursor;

    if (tmp) {
      return { data: tmp.concat(data || []) };
    }

    return { data };
  }

  public async readNextChanges(opts?: ReadOptions): Promise<ReadResult<TRecord>> {
    const limit = opts?.pageSize || this.defaultPageSize;

    if (!this.changeCursor) {
      return { error: Error("cannot fetch changes for this query") };
    }

    const { meta, data, error } = await this.connection.read<TRecord>(this.streamQualifier, { limit, cursor: this.changeCursor });
    if (error) {
      return { error };
    }

    this.changeCursor = meta?.changeCursor;

    return { data };
  }

  public async subscribeChanges(opts: SubscribeOptions<TRecord>): Promise<void> {
    // make sure can subscribe
    if (!this.changeCursor) {
      throw Error("cannot subscribe to changes for this query");
    }
    if (!("instanceID" in this.streamQualifier) || !this.streamQualifier.instanceID) {
      throw Error("cannot subscribe to changes for this query");
    }

    this.connection.subscribe({
      instanceID: this.streamQualifier.instanceID,
      cursor: this.changeCursor,
      onResult: () => {
        // TODO
      },
      onComplete: (error?: Error) => {
        // TODO
      },
    });
  }

}
