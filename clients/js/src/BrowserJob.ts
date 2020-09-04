import { BrowserCursor } from "./BrowserCursor";
import { BrowserConnection, WarehouseJobData } from "./BrowserConnection";

const POLL_FREQUENCY = 1000;

export class BrowserJob<TRecord = any> {
  public jobID?: string;
  public resultAvroSchema?: string;
  public status: "pending" | "running" | "done";
  public referencedInstanceIDs?: string[];
  public bytesScanned?: number;
  public resultSizeBytes?: number;
  public resultSizeRecords?: number;

  private connection: BrowserConnection;
  private jobData: WarehouseJobData;

  constructor(connection: BrowserConnection, jobData: WarehouseJobData) {
    this.connection = connection;
    this.jobData = jobData; // to satisfy init check, overridden by setJobData
    this.status = "pending"; // to satisfy init check, overridden by setJobData
    this.setJobData(jobData);
  }

  public async poll() {
    this.checkIsNotDry();
    const { data, error } = await this.connection.pollWarehouseJob(this.jobID as string);
    if (error || !data) {
      throw Error(`Error polling job: ${error?.message}`);
    }
    this.setJobData(data);
  }

  public async getCursor(): Promise<BrowserCursor<TRecord>> {
    this.checkIsNotDry();
    // poll until completed
    while (this.status !== "done") {
      if (this.jobData) { // don't sleep if we haven't done first fetch yet
        await new Promise(resolve => setTimeout(resolve, POLL_FREQUENCY));
      }
      await this.poll();
    }
    // we know job completed without error (poll raises job errors)
    return new BrowserCursor<TRecord>(this.connection, this.jobData.replayCursor);
  }

  private setJobData(jobData: WarehouseJobData) {
    this.jobData = jobData;
    this.jobID = jobData.jobID;
    this.resultAvroSchema = jobData.resultAvroSchema;
    this.status = jobData.status;
    this.referencedInstanceIDs = jobData.referencedInstanceIDs;
    this.bytesScanned = jobData.bytesScanned;
    this.resultSizeBytes = jobData.resultSizeBytes;
    this.resultSizeRecords = jobData.resultSizeRecords;
  }

  private checkIsNotDry() {
    if (!this.jobID) {
      throw Error("Cannot poll dry run job");
    }
  }

}
