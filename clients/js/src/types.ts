export type Record<TRecord = any> = TRecord & {
  "@meta": { key: string, timestamp: number }
};

export type StreamQualifier = string | { instanceID: string } | { organization: string, project: string, stream: string };
