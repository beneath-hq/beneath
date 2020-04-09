import { BrowserConnection, Response, ResponseMeta } from "./BrowserConnection";
import { DEFAULT_READ_BATCH_SIZE, DEFAULT_SUBSCRIBE_POLL_AT_MOST_EVERY_MS } from "./config";
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

    this.changeCursor = meta?.nextCursor;

    return { data };
  }

  public subscribeChanges(opts: SubscribeOptions<TRecord>): { unsubscribe: () => void } {
    // make sure can subscribe
    if (!this.changeCursor) {
      throw Error("cannot subscribe to changes for this query");
    }
    if ((typeof this.streamQualifier === "string") || !("instanceID" in this.streamQualifier) || !this.streamQualifier.instanceID) {
      throw Error("cannot subscribe to changes for this query");
    }

    const self = this;

    const pageSize = opts.pageSize || DEFAULT_READ_BATCH_SIZE;
    const pollAtMostEveryMilliseconds = opts.pollAtMostEveryMilliseconds || DEFAULT_SUBSCRIBE_POLL_AT_MOST_EVERY_MS;

    const state = {
      stop: false,
      poll: false,
      polling: false,
      lastPoll: 0,
      // tslint:disable-next-line: no-empty
      unsubscribe: () => {},
    };

    const _unsubscribe = (error?: Error) => {
      if (!state.stop) {
        state.stop = true;
        state.unsubscribe();
        opts.onComplete(error);
      }
    };

    const poll = async () => {
      // semantics: Regardless of how many times or how frequently poll() is called, it should make as few
      // network calls as possible (i.e. calls to readNextChanges); however, it must ensure a fetch occurs
      // soon after every call to poll().

      // remember: execution only stops in between awaits; we leverage this extensively

      if (state.stop) { return; }
      state.poll = true;
      if (state.polling) { return; }
      state.polling = true;

      while (state.poll && !state.stop) {
        state.poll = false;

        // ensure poll at most every pollAtMostEveryMilliseconds
        const now = Date.now();
        const delta = Math.max(now - state.lastPoll, 0);
        if (delta < pollAtMostEveryMilliseconds) {
          const sleep = pollAtMostEveryMilliseconds - delta;
          await new Promise(r => setTimeout(r, sleep));
          if (state.stop) { break; }
        }
        state.lastPoll = Date.now();

        const { data, error } = await self.readNextChanges({ pageSize });
        if (state.stop) { break; }

        if (error || !data) {
          _unsubscribe(error);
          break;
        }

        if (data.length > 0) {
          opts.onData(data);
        }

        if (data.length === pageSize) {
          state.poll = true;
        }
      }

      state.polling = false;
    };

    const { unsubscribe } = this.connection.subscribe({
      instanceID: this.streamQualifier.instanceID,
      cursor: this.changeCursor,
      onResult: () => {
        poll();
      },
      onComplete: (error?: Error) => {
        _unsubscribe(error); // idempotent
      },
    });

    state.unsubscribe = unsubscribe;
    return { unsubscribe: _unsubscribe };
  }

}
