import { BrowserClient, BrowserQueryResult } from "beneath";
import { useEffect, useState } from "react";

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
  client?: BrowserClient;
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
  const [client, setClient] = useState<BrowserClient | undefined>(undefined);
  const [records, setRecords] = useState<TRecord[]>([]);
  const [error, setError] = useState<Error | undefined>(undefined);
  const [loading, setLoading] = useState<boolean>(true);
  const [subscribed, setSubscribed] = useState<boolean>(false);
  const [fetchMore, setFetchMore] = useState<FetchMoreFunction | undefined>(undefined);

  // // fetch (and maybe subscribe to) records (will get called every time opts change)
  useEffect(() => {
    // cancellation mechanism for async/await (dirty, but it works)
    // tslint:disable-next-line: max-line-length
    // see: https://dev.to/n1ru4l/homebrew-react-hooks-useasynceffect-or-how-to-handle-async-operations-with-useeffect-1fa8
    let cancel = false;

    // async scope
    (async () => {
      // set default fields in opts for optional fields
      opts = { ...opts }; // clone
      opts.view = opts.view || "lookup";
      opts.pageSize = opts.pageSize || DEFAULT_PAGE_SIZE;
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

      // make client
      const client = new BrowserClient({ secret: opts.secret });
      setClient(client);

      // get stream object
      const stream = client.findStream({
        project: opts.project,
        stream: opts.stream,
        instanceID: opts.instanceID,
      });

      // query stream
      let query: BrowserQueryResult<TRecord>;
      if (opts.view === "latest") {
        query = await stream.peek({ pageSize: opts.pageSize });
      } else if (opts.view === "lookup") {
        query = await stream.query({ compact: true, filter: opts.filter, pageSize: opts.pageSize });
      } else if (opts.view === "log") {
        query = await stream.query({ compact: false, filter: opts.filter, pageSize: opts.pageSize });
      } else {
        throw Error(`invalid view option <${opts.view}>`);
      }
      if (cancel) { return; } // check cancel after await

      // parse query
      if (query.error) {
        setError(query.error);
        return;
      }
      const cursor = query.cursor;
      if (!cursor) {
        throw Error("internal error: expected cursor");
      }

      // fetch first page and set
      const read = await cursor.readNext({ pageSize: opts.pageSize });
      if (cancel) { return; } // check cancel after await
      if (read.error) {
        setError(read.error);
        return;
      }
      let records = read.records || [];
      setRecords(records);

      // done loading
      setLoading(false);

      // set fetch more
      const fetchMore = async (fetchMoreOpts?: FetchMoreOptions) => {
        // stop if cancelled
        if (cancel) { return; }

        // normalize fetchMoreOpts
        fetchMoreOpts = fetchMoreOpts || {};
        fetchMoreOpts.pageSize = fetchMoreOpts.pageSize || opts.pageSize;

        // set loading
        setLoading(true);

        // fetch more
        const read = await cursor.readNext({ pageSize: fetchMoreOpts.pageSize });
        if (cancel) { return; } // check cancel after await
        if (read.error) {
          setError(read.error);
          return;
        }

        // append to records
        records = records.concat(read.records || []);
        setRecords(records);

        // done loading
        setLoading(false);
        return;
      };
      setFetchMore(() => fetchMore);

      // // prepend
      // const prepend = read.records || [];
      // records = prepend.concat(records);
      // setRecords(records);

      //     // done if not asked to subscribe changes
      //     if (!opts.subscribe) {
      //       return;
      //     }

      //     // done if can't subscrine
      //     if (opts.filter) { // || !cursor.canSubscribeChanges
      //       return;
      //     }

      //     // TODO
      //     // establish subscription
      //     // set it to be cancelled in cleanup()
      //     // setSubscribed(true)
      //     // respect pollAtMostEveryMilliseconds and renderAtMostEveryMilliseconds
      //     // on latest: prepend, on log: append, on lookup: merge on @meta.key or insert lexicographically sorted
      //     // stream.subscribeChanges({
      //     //   pollAtMostEveryMilliseconds: opts.subscribe.pollAtMostEveryMilliseconds,
      //     //   onRecords: (records: TRecord[]) => {
      //     //     return;
      //     //   },
      //     // });

      // done with effect
    })();

    return function cleanup() {
      // cancel async/await
      cancel = true;

      // reset state (cleanup is also called when reacting to opts changes), so complete dealloc is not guaranteed
      setFetchMore(undefined);
      setSubscribed(false);
      setLoading(true);
      setError(undefined);
      setRecords([]);
      setClient(undefined);

      // cleanup subscription
    };
  }, [opts.secret, opts.project, opts.stream, opts.instanceID, opts.view, opts.filter, opts.pageSize, opts.subscribe]);

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
