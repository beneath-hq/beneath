export type StreamQualifier = { instanceID: string } | { project: string, stream: string };

export type Record<TRecord = any> = TRecord & {
  "@meta": {
    key: string;
    timestamp: number;
  }
};

export interface ReadResult<TRecord = any> {
  records?: Record<TRecord>[];
  error?: Error;
}

export interface ReadOptions {
  pageSize?: number;
}

export interface QueryOptions extends ReadOptions {
  compact?: boolean;
  filter?: string;
}

