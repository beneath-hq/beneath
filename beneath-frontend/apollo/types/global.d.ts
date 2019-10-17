// GraphQL (control) custom scalars
type ControlUUID = string;
type ControlTime = Date;
type ControlJSON = any;

declare namespace NodeJS {
  export interface Process {
    browser: boolean;
  }
}
