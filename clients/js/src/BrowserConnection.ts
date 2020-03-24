import { BENEATH_GATEWAY_HOST } from "./config";
import { Record, StreamQualifier } from "./shared";

export interface Response<TRecord> {
  records?: Record<TRecord>[];
  error?: Error;
  cursor?: { next?: string, changes?: string };
}

export type QueryArgs = { limit?: number } & ({ compact: boolean, filter?: string } | { cursor: string });

export type PeekArgs = { limit?: number, cursor?: string };

export class BrowserConnection {
  public secret?: string;

  constructor(secret?: string) {
    this.secret = secret;
  }

  public async query<TRecord = any>(streamQualifier: StreamQualifier, args: QueryArgs): Promise<Response<TRecord>> {
    const path = this.makePath(streamQualifier);
    return this.fetch("GET", path, args);
  }

  public async peek<TRecord = any>(streamQualifier: StreamQualifier, args: PeekArgs): Promise<Response<TRecord>> {
    const path = `${this.makePath(streamQualifier)}/peek`;
    return this.fetch("GET", path, args);
  }

  public async write(instanceID: string, records: any[]) {
    // const url = `${connection.GATEWAY_URL}/streams/instances/${instanceID}`;
  }

  private makePath(sq: StreamQualifier): string {
    if ("instanceID" in sq && sq.instanceID) {
      return `streams/instances/${sq.instanceID}`;
    } else if ("project" in sq && "stream" in sq) {
      return `projects/${sq.project}/streams/${sq.stream}`;
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
    const { data, cursor, error } = json;

    if (error) {
      return { error: Error(error) };
    }

    if (data) {
      if (!Array.isArray(data)) {
        throw Error(`Expected array data, got ${JSON.stringify(data)}`);
      }
    }

    return { cursor, records: data };
  }

}
