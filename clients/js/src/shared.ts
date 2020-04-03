export type StreamQualifier = { instanceID: string } | { project: string, stream: string };

export type Record<TRecord = any> = TRecord & {
  "@meta": { key: string, timestamp: number }
};

export type ReadResult<TRecord = any> = { data?: Record<TRecord>[], error?: Error };

export type ReadOptions = { pageSize?: number };

export type QueryLogOptions = ReadOptions & { peek?: boolean; };

export type QueryIndexOptions = ReadOptions & { filter?: string; };

export type SubscribeOptions<TRecord> = {
  onData: (data: Record<TRecord>[]) => void,
  onComplete: (error?: Error) => void,
  pageSize?: number,
  pollAtMostEveryMilliseconds?: number,
};
