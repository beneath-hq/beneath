import { SubscriptionClient } from "subscriptions-transport-ws";

import { BENEATH_GATEWAY_HOST, BENEATH_GATEWAY_HOST_WS } from "./config";
import { Record, StreamQualifier } from "./types";

// Args types

export type PingArgs = { clientID: string, clientVersion: string };

export type ReadArgs = { cursor: string, limit?: number };

export type QueryLogArgs = { limit?: number, peek?: boolean };

export type QueryIndexArgs = { limit?: number, filter?: string };

export type QueryWarehouseArgs = { query: string, dry?: boolean, maxBytesScanned?: number, timeoutMilliseconds?: number };

export type SubscribeArgs = {
  instanceID: string;
  cursor: string;
  onResult: () => void;
  onComplete: (error?: Error) => void;
};

// Response types

export interface Response<Data = any, Meta = any> {
  data?: Data;
  error?: Error;
  meta?: Meta;
}

export type PingData = {
  authenticated: string;
  versionStatus: string;
  recommendedVersion: string;
};

export type RecordsMeta = {
  instanceID?: string;
  nextCursor?: string;
  changeCursor?: string;
};

export type WarehouseJobData = {
  jobID?: string;
  status: "pending" | "running" | "done";
  error?: string;
  resultAvroSchema?: string;
  replayCursor?: string;
  referencedInstanceIDs?: string[];
  bytesScanned?: number;
  resultSizeBytes?: number;
  resultSizeRecords?: number;
};

export type WriteMeta = {
  writeID: string;
};

// Connection class

export class Connection {
  public secret?: string;
  private subscription?: SubscriptionClient;

  constructor(secret?: string) {
    this.secret = secret;

    const connectionParams = secret ? { secret: this.secret } : undefined;
    if (typeof window !== 'undefined') {
      this.subscription = new SubscriptionClient(`${BENEATH_GATEWAY_HOST_WS}/v1/-/ws`, {
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

  public async ping(args: PingArgs): Promise<Response<PingData, null>> {
    const res = await this.fetch<any, null>("GET", "v1/-/ping", {
      client_id: args.clientID,
      client_version: args.clientVersion,
    });

    if (res.data) {
      const data = res.data;
      data.versionStatus = data.version_status;
      delete data.version_status;
      data.recommendedVersion = data.recommended_version;
      delete data.recommended_version;
    }

    return res;
  }

  public async write(streamQualifier: StreamQualifier, records: any[]): Promise<Response<null, WriteMeta>> {
    const path = this.makePath(streamQualifier);
    const res = await this.fetch<null, any>("POST", path, records);

    if (res.meta) {
      res.meta.writeID = res.meta.write_id;
      delete res.meta.write_id;
    }

    return res;
  }

  public async queryLog<TRecord = any>(streamQualifier: StreamQualifier, args: QueryLogArgs): Promise<Response<Record<TRecord>[], RecordsMeta>> {
    const path = this.makePath(streamQualifier);
    return this.fetchRecords(path, { ...args, type: "log" });
  }

  public async queryIndex<TRecord = any>(streamQualifier: StreamQualifier, args: QueryIndexArgs): Promise<Response<Record<TRecord>[], RecordsMeta>> {
    const path = this.makePath(streamQualifier);
    return this.fetchRecords(path, { ...args, type: "index" });
  }

  public async queryWarehouse(args: QueryWarehouseArgs): Promise<Response<WarehouseJobData, null>> {
    const res = await this.fetch<any, null>("POST", "v1/-/warehouse", {
      query: args.query,
      dry: args.dry,
      max_bytes_scanned: args.maxBytesScanned,
      timeout_ms: args.timeoutMilliseconds,
    });
    return this.parseWarehouseResponse(res);
  }

  public async pollWarehouseJob(jobID: string): Promise<Response<WarehouseJobData, null>> {
    const res = await this.fetch<any, null>("GET", `v1/-/warehouse/${jobID}`);
    return this.parseWarehouseResponse(res);
  }

  public async read<TRecord = any>(args: ReadArgs, streamQualifier?: StreamQualifier): Promise<Response<Record<TRecord>[], RecordsMeta>> {
    // we could make all cursor requests to "/v1/-/cursor", but for stream/instance reads it's neat for the URL to show the stream/instance in client logs
    const path = streamQualifier ? this.makePath(streamQualifier) : "v1/-/cursor";
    return this.fetchRecords(path, args);
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

  private async fetch<Data = any, Meta = any>(method: "GET" | "POST", path: string, body?: { [key: string]: any; }): Promise<Response<Data, Meta>> {
    let url = `${BENEATH_GATEWAY_HOST}/${path}?`;
    if (body && method === "GET") {
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
      body: (body && method === "POST") ? JSON.stringify(body) : undefined,
    });

    // if not json: if successful, return empty; else return just error
    if (res.headers.get("Content-Type")?.indexOf("application/json") === -1) {
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

    return { data, meta };
  }

  private async fetchRecords<TRecord>(path: string, body: { [key: string]: any }): Promise<Response<Record<TRecord>[], RecordsMeta>> {
    const res = await this.fetch<Record<TRecord>[], any>("GET", path, body);

    if (res.data) {
      if (!Array.isArray(res.data)) {
        throw Error(`Expected array data, got ${JSON.stringify(res.data)}`);
      }
    }

    if (res.meta) {
      const meta = res.meta;

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

    return res;
  }

  private makePath(sq: StreamQualifier): string {
    if (typeof sq === "string") {
      sq = this.pathStringToObject(sq);
    }
    if ("instanceID" in sq && sq.instanceID) {
      return `v1/-/instances/${sq.instanceID}`;
    } else if ("project" in sq && "stream" in sq) {
      return `v1/${sq.organization}/${sq.project}/${sq.stream}`;
    }
    throw Error("invalid stream qualifier");
  }

  private pathStringToObject(path: string): { organization: string, project: string, stream: string } {
    const parts = path.split("/");

    // trim leading/trailing "/"
    if (parts.length > 0 && parts[0] === "/") { parts.shift(); }
    if (parts.length > 0 && parts[parts.length - 1] === "/") { parts.pop(); }

    // handle org/proj/stream
    if (parts.length === 3) {
      return {
        organization: parts[0],
        project: parts[1],
        stream: parts[2],
      };
    }

    throw Error(`Cannot parse stream path "${path}"; it must have the format "organization/project/stream"`);
  }

  private parseWarehouseResponse(res: Response<any, null>): Response<WarehouseJobData, null> {
    if (res.error || !res.data) {
      return res;
    }

    const data: WarehouseJobData = {
      jobID: res.data.job_id,
      status: res.data.status,
      error: res.data.error,
      resultAvroSchema: res.data.result_avro_schema,
      replayCursor: res.data.replay_cursor,
      referencedInstanceIDs: res.data.referenced_instances,
      bytesScanned: res.data.bytes_scanned,
      resultSizeBytes: res.data.result_size_bytes,
      resultSizeRecords: res.data.result_size_records,
    };

    return { data };
  }

}
