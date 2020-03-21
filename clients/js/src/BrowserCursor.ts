import { BrowserConnection } from "./BrowserConnection";
import { Record, ReadOptions, ReadResult, StreamQualifier } from "./shared";

export class BrowserCursor<TRecord = any> {
  private connection: BrowserConnection;
  private streamQualifier: StreamQualifier;
  private nextCursor?: string;
  private changesCursor?: string;
  private buffer?: Record<TRecord>[];


  constructor(connection: BrowserConnection, streamQualifier: StreamQualifier, initialRecords: Record<TRecord>[], nextCursor?: string, changesCursor?: string) {
    this.connection = connection;
    this.streamQualifier = streamQualifier;
    this.nextCursor = nextCursor;
    this.changesCursor = changesCursor;
    this.buffer = initialRecords;
  }

  public async readNext(opts?: ReadOptions): Promise<ReadResult<TRecord>> {
    return { records: this.buffer };
  }

}
