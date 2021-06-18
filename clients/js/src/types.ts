export type Record<TRecord = any> = TRecord & {
  "@meta": { key: string, timestamp: number }
};

export type TableQualifier = string | { instanceID: string } | { organization: string, project: string, table: string };
