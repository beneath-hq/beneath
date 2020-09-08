export type FetchMoreOptions = { pageSize?: number };
export type FetchMoreFunction = (opts?: FetchMoreOptions) => Promise<void>;
