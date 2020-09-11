import { Client, Record, Job } from "beneath";
import { v4 as uuidv4 } from "uuid";
import { useState } from "react";

import { FetchMoreFunction, FetchMoreOptions } from "./shared";

export type QueryWarehouseFunction = (query: string) => void;
export type UseWarehouseOptions = { secret?: string, pageSize?: number };
export type UseWarehouseResult<TRecord> = {
  client: Client,
  analyzeQuery: QueryWarehouseFunction,
  runQuery: QueryWarehouseFunction,
  loading: boolean,
  error?: Error,
  job?: Job<TRecord>,
  records?: Record<TRecord>[],
  fetchMore?: FetchMoreFunction,
};

const DEFAULT_PAGE_SIZE = 25;

export function useWarehouse<TRecord = any>(opts: UseWarehouseOptions): UseWarehouseResult<TRecord> {
  opts.pageSize = opts.pageSize ?? DEFAULT_PAGE_SIZE;

  const [client] = useState<Client>(() => new Client({ secret: opts.secret }));
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<Error | undefined>(undefined);
  const [job, setJob] = useState<Job | undefined>(undefined);
  const [data, setData] = useState<{ records?: Record<TRecord>[] }>({ records: undefined });
  const [fetchMore, setFetchMore] = useState<FetchMoreFunction | undefined>(undefined);

  const startQuery = async (query: string, dry: boolean) => {
    setLoading(true);
    setError(undefined);
    setJob(undefined);
    setData({});
    setFetchMore(undefined);

    const stop = (error?: Error) => {
      setError(error);
      setLoading(false);
      setFetchMore(undefined);
    };

    const { job, error } = await client.queryWarehouse<TRecord>({ query, dry });
    if (error || !job) {
      return stop(error);
    }
    setJob(job);

    if (dry) {
      return stop();
    }

    const onPoll = () => setJob(job); // trigger refresh when job state changes
    const { cursor, error: cursorError } = await job.getCursor({ onPoll });
    if (cursorError || !cursor) {
      return stop(cursorError);
    }

    if (!cursor.hasNext()) {
      setData({ records: [] });
      return stop();
    }

    const fetchMore = async (opts?: FetchMoreOptions) => {
      setLoading(true);
      const { data: _batch, error } = await cursor.readNext({ pageSize: opts?.pageSize });
      if (error) {
        return stop(error);
      }
      const batch = _batch ?? [];

      // add key to batch
      for (const record of batch) {
        if (!record["@meta"].key) {
          record["@meta"].key = uuidv4();
        }
      }

      setData((data) => {
        const records = data.records ? data.records.concat(batch) : batch;
        return { records };
      });
      setLoading(false);

      if (!cursor.hasNext()) {
        setFetchMore(undefined);
      }
    };

    await fetchMore({ pageSize: opts.pageSize });
    if (cursor.hasNext()) {
      setFetchMore(() => fetchMore);
    }
  };

  return {
    client,
    analyzeQuery: (query: string) => startQuery(query, true),
    runQuery: (query: string) => startQuery(query, false),
    loading,
    error,
    job,
    records: data.records,
    fetchMore,
  };
}
