import { BENEATH_GATEWAY_HOST } from "./config";
import { Record, StreamQualifier } from "./shared";

export interface Response<TRecord> {
  records?: Record<TRecord>[];
  error?: Error;
  cursor?: { next?: string, changes?: string };
}

export class BrowserConnection {
  public secret?: string;

  constructor(secret?: string) {
    this.secret = secret;
  }

  public async query<TRecord = any>(streamQualifier: StreamQualifier, compact: boolean, filter?: string, pageSize?: number): Promise<Response<TRecord>> {
    let path;
    if ("instanceID" in streamQualifier) {
      path = `streams/instances/${streamQualifier.instanceID}`;
    } else {
      path = `projects/${streamQualifier.project}/streams/${streamQualifier.stream}`;
    }

    const args = { compact, filter, limit: pageSize };

    return this.fetch("GET", path, args);
  }

  public async peek<TRecord = any>(streamQualifier: StreamQualifier, pageSize?: number): Promise<Response<TRecord>> {
    let path;
    if ("instanceID" in streamQualifier) {
      path = `streams/instances/${streamQualifier.instanceID}/latest`;
    } else {
      path = `projects/${streamQualifier.project}/streams/${streamQualifier.stream}/latest`;
    }

    return this.fetch("GET", path, { limit: pageSize });
  }

  public async write(instanceID: string, records: any[]) {
    // const url = `${connection.GATEWAY_URL}/streams/instances/${instanceID}`;
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

    // parse as json
    let json;
    try {
      json = await res.json();
    } catch {
      // not json: if successful, return empty; else return just error
      if (res.ok) {
        return {};
      }
      let error = await res.text();
      if (!error) {
        error = res.statusText;
      }
      return { error: Error(error) };
    }

    // check json and return
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
