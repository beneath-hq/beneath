import { Client, StreamQualifier } from "beneath";
import { sortedUniqBy } from "lodash";
import { useEffect, useState } from "react";

import { FetchMoreFunction, FetchMoreOptions } from "./shared";

const DEFAULT_PAGE_SIZE = 25;
const DEFAULT_FLASH_DURATION = 2000;
const DEFAULT_RENDER_FREQUENCY = 250;
const DEFAULT_SUBSCRIBE_POLL_FREQUENCY = 250;

/**
 * Options passed to {@linkcode useRecords}
 */
export interface UseRecordsOptions {
  /** Secret to use for authentication */
  secret?: string;
  /** Identifier for the stream or instance to query */
  stream: StreamQualifier;
  /** Query to run. Defaults to an unfiltered index query. */
  query?: { type: "index", filter?: string } | { type: "log", peek?: boolean };
  /** Number of records to fetch per request (limit). Defaults to 25. */
  pageSize?: number;
  /** If configured, will open a websocket and trigger a re-render when new records are received. */
  subscribe?: boolean | { pageSize?: number, pollFrequencyMs?: number };
  /** If set, will truncate results to `maxRecords` by removing records */
  truncatePolicy?: "start" | "end" | "auto";
  /** Max number of records to keep in memory before applying `truncatePolicy` */
  maxRecords?: number;
  /** If subscribed to websockets, sets the duration that each new record will have `record["@meta"].flash === true` */
  flashDurationMs?: number;
  /** If subscribed to websockets, sets the max frequency at which re-renders are triggered (to prevent lagging) */
  renderFrequencyMs?: number;
}

export interface UseRecordsResult<TRecord> {
  /** Client used to connect (see the non-react beneath library) */
  client?: Client;
  /** Records returned by query. If subscribed to websockets, may change on re-renders */
  records: Record<TRecord>[];
  /** Error returned by the query */
  error?: Error;
  /** Callback to trigger fetching another page */
  fetchMore?: FetchMoreFunction;
  /** Callback to trigger fetching changes (not applicable if subscribed to websockets) */
  fetchMoreChanges?: FetchMoreFunction;
  /** True if currently loading for the first time (not if subscribed to websockets) */
  loading: boolean;
  /** Status on the websockets subscription */
  subscription: {
    /** True if a websocket connection is open */
    online: boolean;
    /** Error returned from the websocket connection */
    error?: Error;
  };
  /** Status on whether records have been truncated under the truncation policy */
  truncation: {
    /** True if records were truncated from the start */
    start: boolean;
    /** True if records were truncated from the start */
    end: boolean;
  };
}

// like beneath.Record, but with @meta.flash added
export type Record<TRecord = any> = TRecord & {
  "@meta": {
    key: string;
    timestamp: number;
    flash?: boolean;
    _key?: Buffer;
    _flashTime?: number;
  };
};

/**
 * React hook that you can use to query streams, including paging through data
 * and getting real-time updates over websockets.
 * @param opts  Options, including required parameters. See {@linkcode UseRecordsOptions} for details.
 */
export function useRecords<TRecord = any>(opts: UseRecordsOptions): UseRecordsResult<TRecord> {
  // values
  const [client, setClient] = useState<Client>(() => new Client({ secret: opts.secret }));
  const [data, setData] = useState<{ records: Record<TRecord>[] }>({ records: [] });
  const [error, setError] = useState<Error | undefined>(undefined);
  const [fetchMore, setFetchMore] = useState<FetchMoreFunction | undefined>(undefined);
  const [fetchMoreChanges, setFetchMoreChanges] = useState<FetchMoreFunction | undefined>(undefined);

  // flags
  const [loading, setLoading] = useState<boolean>(true);
  const [truncatedStart, setTruncatedStart] = useState<boolean>(false);
  const [truncatedEnd, setTruncatedEnd] = useState<boolean>(false);
  const [subscriptionOnline, setSubscriptionOnline] = useState<boolean>(false);
  const [subscriptionError, setSubscriptionError] = useState<Error | undefined>(undefined);

  // // fetch (and maybe subscribe to) records (will get called every time opts change)
  useEffect(() => {
    // Section: PARSING OPTIONS

    const queryType = opts.query?.type || "index";
    const queryPeek = opts.query?.type === "log" ? opts.query.peek || false : undefined;
    const queryFilter = opts.query?.type === "index" ? opts.query.filter : undefined;
    const pageSize = opts.pageSize || DEFAULT_PAGE_SIZE;
    const maxRecords = opts.maxRecords;
    const truncatePolicy = opts.truncatePolicy || "auto";
    const flashDuration = opts.flashDurationMs === undefined ? DEFAULT_FLASH_DURATION : opts.flashDurationMs;
    const renderFrequency = opts.renderFrequencyMs || DEFAULT_RENDER_FREQUENCY;

    const subscribeEnabled = typeof window === "undefined" ? false : !!opts.subscribe;
    const subscribeOpts = (typeof opts.subscribe === "object") ? opts.subscribe : {};
    const subscribePageSize = subscribeOpts.pageSize || pageSize;
    const subscribePollFrequency = subscribeOpts.pollFrequencyMs || DEFAULT_SUBSCRIBE_POLL_FREQUENCY;

    if (!maxRecords && opts.truncatePolicy) {
      throw Error("useRecords cannot apply a truncate policy when maxRecords is not specified");
    }

    // SECTION: State stored outside async scope

    // cancellation mechanism for async/await (dirty, but it works)
    // tslint:disable-next-line: max-line-length
    // see: https://dev.to/n1ru4l/homebrew-react-hooks-useasynceffect-or-how-to-handle-async-operations-with-useeffect-1fa8
    let cancel = false;

    let loadedMore = false;
    let subscriptionUnsubscribe: (() => any) | undefined;

    // async scope
    (async () => {

      // Section: INITIAL LOAD

      // get stream object
      const stream = client.findStream(opts.stream);

      // query stream
      const query =
        queryType === "index"
          ? await stream.queryIndex({ filter: queryFilter, pageSize })
          : queryType === "log"
            ? await stream.queryLog({ pageSize, peek: queryPeek })
            : undefined;
      if (!query) {
        throw Error(`invalid view option <${queryType}>`);
      }
      if (cancel) { return; } // check cancel after await

      // parse query
      if (query.error) {
        setError(query.error);
        setLoading(false);
        return;
      }
      const cursor = query.cursor;
      if (!cursor) {
        throw Error("internal error: expected cursor");
      }

      // fetch first page and set
      if (cursor.hasNext()) {
        const read = await cursor.readNext({ pageSize });
        if (cancel) { return; } // check cancel after await
        if (read.error) {
          setError(read.error);
          setLoading(false);
          return;
        }
        setData({ records: read.data || [] });
      }

      // done loading
      setLoading(false);

      // Section: BUFFERED RENDERING LOGIC

      let lastRender = 0;
      let renderScheduled = false;
      let deflashScheduled = false;
      let moreBuffer: Record<TRecord>[] = [];
      let changesBuffer: Record<TRecord>[] = [];

      const deflash = (records: Record<TRecord>[]) => {
        const cutoff = Date.now() - flashDuration;
        for (const record of records) {
          if (record["@meta"]._flashTime && record["@meta"]._flashTime <= cutoff) {
            delete record["@meta"].flash;
            delete record["@meta"]._flashTime;
          }
        }
        return records;
      };

      const render = () => {
        if (cancel) { return; }

        // update records
        let data: undefined | { records: Record<TRecord>[] };
        setData((currentData) => {
          data = currentData;
          return currentData;
        });

        // NOTE: This manoeuvre to get a fresh copy of data is horrible! It only works because there's
        // no async code before the next call to setData. Why not just put the whole block below in the
        // setData callback above? Because the setXXX calls in the code somehow triggered a nested rerun
        // of the setData callback precisely on the 20th run. This is such a weird bug, and this horrible
        // hack (seemingly) solves it.

        if (!data) {
          throw Error("unexpected internal error!");
        }

        let result = data.records;
        if (deflashScheduled) {
          result = deflash(result);
        }

        if (moreBuffer.length > 0) {
          result.push(...moreBuffer);
        }

        if (changesBuffer.length > 0) {
          if (queryType === "index") {
            result = mergeIndexedRecordsAndChanges(result, changesBuffer, false, cursor.hasNext());
          } else if (queryType === "log") {
            if (queryPeek) {
              result.unshift(...changesBuffer.reverse());
            } else {
              result.push(...changesBuffer);
            }
          }
        }

        if (maxRecords && result.length > maxRecords) {
          if (truncatePolicy === "start") {
            result = result.slice(result.length - maxRecords);
            setTruncatedStart(true);
          } else if (truncatePolicy === "end") {
            result = result.slice(0, maxRecords);
            setTruncatedEnd(true);
          } else if (truncatePolicy === "auto") {
            if (queryType === "log" && !queryPeek) {
              // truncate start
              result = result.slice(result.length - maxRecords);
              setTruncatedStart(true);
            } else {
              if (loadedMore) {
                // truncate start
                result = result.slice(result.length - maxRecords);
                setTruncatedStart(true);

                // disable subscription (since the user is likely paging anyway)
                if (subscriptionUnsubscribe) {
                  subscriptionUnsubscribe();
                }
                setSubscriptionError(Error("Real-time updates were disabled since too many rows have been loaded (refresh page to reset)"));
                setSubscriptionOnline(false);
              } else {
                // truncate end
                result = result.slice(0, maxRecords);
                setTruncatedEnd(true);

                // disable fetching more (since ends wouldn't match)
                setFetchMore(undefined);
              }
            }
          }
        }

        setData({ records: result });

        lastRender = Date.now();
        renderScheduled = false;
        deflashScheduled = false;
        moreBuffer = [];
        changesBuffer = [];
      };

      const scheduleRender = (deflash: boolean) => {
        if (deflash) {
          deflashScheduled = true;
        }
        if (!renderScheduled) {
          renderScheduled = true;
          const delta = Math.max(0, Date.now() - lastRender);
          const wait = renderFrequency - delta;
          if (wait > 0) {
            setTimeout(render, wait);
          } else {
            render();
          }
        }
      };

      const handleFlash = (data: Record<TRecord>[], flash: boolean) => {
        flash = flash && flashDuration !== 0;
        if (flash) {
          const now = Date.now();
          // tslint:disable-next-line: prefer-for-of
          for (let i = 0; i < data.length; i++) {
            const record = data[i] as Record<TRecord>;
            record["@meta"].flash = true;
            record["@meta"]._flashTime = now;
          }

          // schedule deflash
          setTimeout(() => {
            if (cancel) { return; }
            scheduleRender(true);
          }, flashDuration);
        }
      };

      const handleChangeData = (data: Record<TRecord>[], flash: boolean) => {
        handleFlash(data, flash);
        changesBuffer = changesBuffer.concat(data);
        scheduleRender(false);
      };

      const handleMoreData = (data: Record<TRecord>[], flash: boolean) => {
        handleFlash(data, flash);
        moreBuffer = moreBuffer.concat(data);
        scheduleRender(false);
      };

      // SECTION: Fetching more

      // set fetch more
      if (cursor.hasNext()) {
        const fetchMore = async (fetchMoreOpts?: FetchMoreOptions) => {
          // stop if cancelled
          if (cancel) { return; }

          // normalize fetchMoreOpts
          const fetchMorePageSize = fetchMoreOpts?.pageSize || pageSize;

          // set loading
          setLoading(true);

          // update flag used for 'auto' truncate policy
          loadedMore = true;

          // fetch more
          const read = await cursor.readNext({ pageSize: fetchMorePageSize });
          if (cancel) { return; } // check cancel after await
          if (read.error) {
            setError(read.error);
            setLoading(false);
            return;
          }

          // append to records
          handleMoreData(read.data || [], true);

          // update fetch more
          if (!cursor.hasNext()) {
            setFetchMore(undefined);
          }

          // done loading
          setLoading(false);
          return;
        };
        setFetchMore(() => fetchMore);
      }

      // SECTION: Changes
      if (cursor.hasNextChanges() && !queryFilter) {
        if (subscribeEnabled && !(queryType === "log" && !queryPeek && cursor.hasNext())) {
          // create subscription
          const { unsubscribe } = cursor.subscribeChanges({
            pageSize: subscribePageSize,
            pollAtMostEveryMilliseconds: subscribePollFrequency,
            onData: (data) => {
              handleChangeData(data, true);
            },
            onComplete: (error) => {
              if (cancel) { return; }
              setSubscriptionOnline(false);
              setSubscriptionError(error);
            },
          });
          subscriptionUnsubscribe = unsubscribe;
          setSubscriptionOnline(true);
        } else {
          // create manual fetch changes handler
          const fetchMoreChanges = async (fetchMoreOpts?: FetchMoreOptions) => {
            // stop if cancelled
            if (cancel) { return; }

            // normalize fetchMoreOpts
            const fetchMorePageSize = fetchMoreOpts?.pageSize || pageSize;

            // set loading
            setLoading(true);

            // fetch more
            const read = await cursor.readNextChanges({ pageSize: fetchMorePageSize });
            if (cancel) { return; } // check cancel after await
            if (read.error) {
              setError(read.error);
              setLoading(false);
              return;
            }

            // append to records
            handleChangeData(read.data || [], true);

            // update fetch more
            if (!cursor.hasNextChanges()) {
              setFetchMoreChanges(undefined);
            }

            // done loading
            setLoading(false);
            return;
          };
          setFetchMoreChanges(() => fetchMoreChanges);
        }

      }

      // done with effect
    })();

    return function cleanup() {
      // cancel async/await
      cancel = true;

      // stop subscription if open
      if (subscriptionUnsubscribe) {
        subscriptionUnsubscribe();
      }

      // reset state (cleanup is also called when reacting to opts changes), so complete dealloc is not guaranteed
      setSubscriptionError(undefined);
      setSubscriptionOnline(false);
      setTruncatedEnd(false);
      setTruncatedStart(false);
      setLoading(true);
      setFetchMoreChanges(undefined);
      setFetchMore(undefined);
      setError(undefined);
      setData({ records: [] });

      // cleanup subscription
    };
  }, [
    opts.secret,
    (typeof opts.stream === "string") ? opts.stream : ("instanceID" in opts.stream) ? opts.stream.instanceID : `${opts.stream.organization}/${opts.stream.project}/${opts.stream.stream}`,
    opts.query?.type,
    opts.query?.type === "log" ? opts.query.peek : opts.query?.type === "index" ? opts.query.filter : undefined,
    opts.pageSize,
    !!opts.subscribe,
  ]);

  // UseRecordsResult
  return {
    client,
    records: data.records,
    error,
    loading,
    fetchMore,
    fetchMoreChanges,
    subscription: {
      online: subscriptionOnline,
      error: subscriptionError,
    },
    truncation: {
      start: truncatedStart,
      end: truncatedEnd,
    },
  };
}

function mergeIndexedRecordsAndChanges<TRecord>(
  records: Record<TRecord>[],
  changes: Record<TRecord>[],
  skipStart: boolean,
  skipEnd: boolean,
): Record<TRecord>[] {
  // sort changes a) lexicographically by _key, b) equal keys sorted descending by timestamp
  changes.sort((a, b) => {
    const ak = getBufferKey(a);
    const bk = getBufferKey(b);
    const c = ak.compare(bk);
    if (c === 0) {
      const at = a["@meta"].timestamp;
      const bt = b["@meta"].timestamp;
      if (at < bt) {
        return 1;
      } else if (at > bt) {
        return -1;
      } else {
        return 0;
      }
    }
    return c;
  });

  // remove duplicates (keeps order, keeps first occurence)
  const sortedUniqChanges = sortedUniqBy(changes, (record) => record["@meta"].key);

  // construuct new merged array
  const result = [];
  let recordsIdx = 0;
  let changesIdx = 0;

  // iterate and merge
  while (changesIdx < sortedUniqChanges.length || recordsIdx < records.length) {
    if (changesIdx < sortedUniqChanges.length && recordsIdx < records.length) {
      const changeRecord = sortedUniqChanges[changesIdx];
      const existingRecord = records[recordsIdx];

      const changeKey = getBufferKey(changeRecord);
      const existingKey = getBufferKey(existingRecord);

      const comp = changeKey.compare(existingKey);
      if (comp < 0) {
        if (!skipStart || recordsIdx !== 0) {
          result.push(changeRecord);
        }
        changesIdx++;
      } else if (comp > 0) {
        result.push(existingRecord);
        recordsIdx++;
      } else {
        if (changeRecord["@meta"].timestamp >= existingRecord["@meta"].timestamp) {
          result.push(changeRecord);
        } else {
          result.push(existingRecord);
        }
        changesIdx++;
        recordsIdx++;
      }
    } else if (recordsIdx < records.length) {
      result.push(records[recordsIdx]);
      recordsIdx++;
    } else if (changesIdx < sortedUniqChanges.length) {
      if (skipEnd) {
        changesIdx = sortedUniqChanges.length;
      } else {
        result.push(sortedUniqChanges[changesIdx]);
        changesIdx++;
      }
    }
  }

  // done
  return result;
}

function getBufferKey<TRecord>(record: Record<TRecord>): Buffer {
  if (record["@meta"]._key === undefined) {
    record["@meta"]._key = Buffer.from(record["@meta"].key, "base64");
  }
  return record["@meta"]._key;
}
