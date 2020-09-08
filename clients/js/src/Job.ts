import { Cursor } from "./Cursor";
import { Connection, WarehouseJobData } from "./Connection";

const POLL_FREQUENCY = 1000;

export type GetCursorOptions = { onPoll: () => void };

export class Job<TRecord = any> {
  public jobID?: string;
  public status: "pending" | "running" | "done";
  public resultAvroSchema?: string;
  public replayCursor?: string;
  public referencedInstanceIDs?: string[];
  public bytesScanned?: number;
  public resultSizeBytes?: number;
  public resultSizeRecords?: number;

  private connection: Connection;
  private jobData: WarehouseJobData;

  constructor(connection: Connection, jobData: WarehouseJobData) {
    this.connection = connection;
    this.jobData = jobData; // to satisfy init check, overridden by setJobData
    this.status = "pending"; // to satisfy init check, overridden by setJobData
    this.setJobData(jobData);
  }

  public async poll(): Promise<{ error?: Error }> {
    this.checkIsNotDry();
    const { data, error } = await this.connection.pollWarehouseJob(this.jobID as string);
    if (error) {
      return { error };
    }
    if (!data) {
      throw Error(`Error data is undefined`);
    }
    this.setJobData(data);
    return {};
  }

  public async getCursor(opts?: GetCursorOptions): Promise<{ cursor?: Cursor<TRecord>, error?: Error }> {
    this.checkIsNotDry();
    // poll until completed
    while (this.status !== "done") {
      if (this.jobData) { // don't sleep if we haven't done first fetch yet
        await new Promise(resolve => setTimeout(resolve, POLL_FREQUENCY));
      }
      const { error } = await this.poll();
      if (error) {
        return { error };
      }
      if (opts?.onPoll) {
        opts.onPoll();
      }
    }
    // we know job completed without error
    const cursor = new Cursor<TRecord>(this.connection, this.jobData.replayCursor);
    return { cursor };
  }

  private setJobData(jobData: WarehouseJobData) {
    this.jobData = jobData;
    this.jobID = jobData.jobID;
    this.status = jobData.status;
    this.resultAvroSchema = jobData.resultAvroSchema ?? this.resultAvroSchema;
    this.replayCursor = jobData.replayCursor;
    this.referencedInstanceIDs = jobData.referencedInstanceIDs ?? this.referencedInstanceIDs;
    this.bytesScanned = jobData.bytesScanned ?? this.bytesScanned;
    this.resultSizeBytes = jobData.resultSizeBytes ?? this.resultSizeBytes;
    this.resultSizeRecords = jobData.resultSizeRecords ?? this.resultSizeRecords;
  }

  private checkIsNotDry() {
    if (!this.jobID) {
      throw Error("Cannot poll dry run job");
    }
  }

}
