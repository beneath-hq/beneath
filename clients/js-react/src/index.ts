import { BrowserClient } from "beneath";
import { useState, useEffect, FC } from "react";
import { SubscriptionClient } from "subscriptions-transport-ws";

const DEFAULT_PAGE_SIZE = 25;
const DEFAULT_SUBSCRIBE_POLL_AT_MOST_EVERY_MILLISECONDS = 250;
const DEFAULT_SUBSCRIBE_RENDER_AT_MOST_EVERY_MILLISECONDS = 250;

export interface UseRecordsOptions {
  secret?: string;
  project: string;
  stream: string;
  instanceID?: string;
  view?: "lookup" | "log" | "latest";
  filter?: string;
  pageSize?: number;
  subscribe?: boolean | SubscribeOptions;
}

export interface UseRecordsResult<TRecord> {
  client: BrowserClient;
  records: TRecord[];
  error?: Error;
  loading: boolean;
  subscribed: boolean;
  fetchMore?: FetchMoreFunction;
}

export interface SubscribeOptions {
  pollAtMostEveryMilliseconds?: number;
  renderAtMostEveryMilliseconds?: number;
}

export type FetchMoreFunction = (opts?: FetchMoreOptions) => Promise<void>;

export interface FetchMoreOptions {
  pageSize?: number;
}

export function useRecords<TRecord = any>(opts: UseRecordsOptions): UseRecordsResult<TRecord> {
  // prepare state
  const [client, setClient] = useState<BrowserClient>(() => new BrowserClient({ secret: opts.secret }));
  const [records, setRecords] = useState<TRecord[]>([]);
  const [error, setError] = useState<Error | undefined>(undefined);
  const [loading, setLoading] = useState<boolean>(true);
  const [subscribed, setSubscribed] = useState<boolean>(false);
  const [fetchMore, setFetchMore] = useState<FetchMoreFunction | undefined>(undefined);

  // set default fields in opts for optional fields
  if (!opts.view) {
    opts.view = "lookup";
  }
  if (!opts.pageSize) {
    opts.pageSize = DEFAULT_PAGE_SIZE;
  }
  if (opts.subscribe) {
    if (typeof opts.subscribe === "boolean") {
      opts.subscribe = {};
    }
    if (!opts.subscribe.pollAtMostEveryMilliseconds) {
      opts.subscribe.pollAtMostEveryMilliseconds = DEFAULT_SUBSCRIBE_POLL_AT_MOST_EVERY_MILLISECONDS;
    }
    if (!opts.subscribe.renderAtMostEveryMilliseconds) {
      opts.subscribe.renderAtMostEveryMilliseconds = DEFAULT_SUBSCRIBE_RENDER_AT_MOST_EVERY_MILLISECONDS;
    }
  }

  // fetch (and maybe subscribe to) records (will get called every time opts change)
  useEffect(() => {
    /*
    TODO (look at BeneathAPI):
    - get stream
    - run query
    - set records
    - if possible, create subscription
    - on subscription result, merge on key if lookup, else prepend/append
      - if result is insert, bad luck until server-side handled
    */

    return function cleanup() {
      // TODO: cleanup everything (records, error, subscriptions, fetchcMore, etc.)
    };
  }, [opts]);

  // UseRecordsResult
  return {
    client,
    records,
    error,
    loading,
    subscribed,
    fetchMore,
  };
}
