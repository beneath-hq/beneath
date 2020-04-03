import { BENEATH_GATEWAY_HOST, BENEATH_GATEWAY_HOST_WS } from "./config";
import { Record, StreamQualifier } from "./shared";
import { SubscriptionClient } from "subscriptions-transport-ws";

export interface Response<TRecord> {
  data?: Record<TRecord>[];
  error?: Error;
  meta?: ResponseMeta;
}

export type ResponseMeta = {
  instanceID?: string;
  nextCursor?: string;
  changeCursor?: string;
};

export type QueryLogArgs = { limit?: number, peek?: boolean };

export type QueryIndexArgs = { limit?: number, filter?: string };

export type ReadArgs = { cursor: string, limit?: number };

export type SubscribeArgs = {
  instanceID: string;
  cursor: string;
  onResult: () => void;
  onComplete: (error?: Error) => void;
};

export class BrowserConnection {
  public secret?: string;
  private subscription?: SubscriptionClient;

  constructor(secret?: string) {
    this.secret = secret;

    const connectionParams = secret ? { secret: this.secret } : undefined;
    if (typeof window !== 'undefined') {
      this.subscription = new SubscriptionClient(`${BENEATH_GATEWAY_HOST_WS}/v1/ws`, {
        lazy: true,
        inactivityTimeout: 10000,
        reconnect: true,
        connectionParams,
        connectionCallback: (error: Error[], result?: any) => {
          if (error) {
            console.error("Beneath subscription error: ", error);
          }
        },
      });
    }
  }

  public async queryLog<TRecord = any>(streamQualifier: StreamQualifier, args: QueryLogArgs): Promise<Response<TRecord>> {
    const path = this.makePath(streamQualifier);
    return this.fetch("GET", path, { ...args, type: "log" });
  }

  public async queryIndex<TRecord = any>(streamQualifier: StreamQualifier, args: QueryIndexArgs): Promise<Response<TRecord>> {
    const path = this.makePath(streamQualifier);
    return this.fetch("GET", path, { ...args, type: "index" });
  }

  public async read<TRecord = any>(streamQualifier: StreamQualifier, args: ReadArgs): Promise<Response<TRecord>> {
    const path = `${this.makePath(streamQualifier)}`;
    return this.fetch("GET", path, args);
  }

  public async write(instanceID: string, records: any[]) {
    // const url = `${connection.GATEWAY_URL}/streams/instances/${instanceID}`;
    // TODO
  }

  public subscribe<TRecord = any>(args: SubscribeArgs): { unsubscribe: () => void } {
    if (!this.subscription) {
      throw Error("cannot subscribe to websocket updates outside of browser");
    }
    const payload = { query: " ", cursor: args.cursor, instance_id: args.instanceID };
    const req = this.subscription.request(payload).subscribe({
      next: (result) => {
        args.onResult();
      },
      error: (error) => {
        req.unsubscribe();
        args.onComplete(error);
      },
      complete: () => {
        args.onComplete();
      },
    });
    return { unsubscribe: req.unsubscribe };
  }

  private makePath(sq: StreamQualifier): string {
    if ("instanceID" in sq && sq.instanceID) {
      return `v1/streams/instances/${sq.instanceID}`;
    } else if ("project" in sq && "stream" in sq) {
      return `v1/projects/${sq.project}/streams/${sq.stream}`;
    }
    throw Error("invalid stream qualifier");
  }

  private async fetch<TRecord>(method: "GET" | "POST", path: string, body: { [key: string]: any; }): Promise<Response<TRecord>> {
    let url = `${BENEATH_GATEWAY_HOST}/${path}?`;
    if (method === "GET") {
      for (const key of Object.keys(body)) {
        const val = body[key];
        if (val !== undefined) {
          url += `${key}=${val}&`;
        }
      }
    }

    const headers: any = { "Content-Type": "application/json" };
    if (this.secret) {
      headers.Authorization = `Bearer ${this.secret}`;
    }

    const res = await fetch(url, {
      method,
      headers,
      body: method === "POST" ? JSON.stringify(body) : undefined,
    });

    // if not json: if successful, return empty; else return just error
    if (res.headers.get("Content-Type") !== "application/json") {
      if (res.ok) {
        return {};
      }
      let error = await res.text();
      if (!error) {
        error = res.statusText;
      }
      return { error: Error(error) };
    }

    // parse json and return
    const json = await res.json();
    const { data, meta, error } = json;

    if (error) {
      return { error: Error(error) };
    }

    if (data) {
      if (!Array.isArray(data)) {
        throw Error(`Expected array data, got ${JSON.stringify(data)}`);
      }
    }

    if (meta) {
      if (meta.instance_id) {
        meta.instanceID = meta.instance_id;
        delete meta.instance_id;
      }

      if (meta.next_cursor) {
        meta.nextCursor = meta.next_cursor;
        delete meta.next_cursor;
      }

      if (meta.change_cursor) {
        meta.changeCursor = meta.change_cursor;
        delete meta.change_cursor;
      }
    }

    return { data, meta };
  }

}
