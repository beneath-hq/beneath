import { BrowserCursor } from "./BrowserCursor";
import { BrowserConnection, WarehouseJobData } from "./BrowserConnection";

export class BrowserJob<TRecord = any> {
  private connection: BrowserConnection;
  private jobData: WarehouseJobData;

  constructor(connection: BrowserConnection, jobData: WarehouseJobData) {
    this.connection = connection;
    this.jobData = jobData;
  }

  public async poll() {
    // TODO:
    throw Error("not implemented");
  }

  public async getCursor(): Promise<BrowserCursor<TRecord>> {
    // TODO:
    throw Error("not implemented");
  }

}
