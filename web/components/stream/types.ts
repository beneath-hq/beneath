export interface StreamInstance {
  streamInstanceID: string;
  version: number;
  createdOn: ControlTime;
  madePrimaryOn: ControlTime | null;
  madeFinalOn: ControlTime | null;
}
