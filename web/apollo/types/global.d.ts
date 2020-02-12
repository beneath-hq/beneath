// GraphQL (control) custom scalars
type ControlUUID = string;
type ControlTime = string;
type ControlJSON = any;

declare namespace NodeJS {
  export interface Process {
    browser: boolean;
  }
}
