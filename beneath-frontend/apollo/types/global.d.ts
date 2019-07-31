// GraphQL (control) custom scalars
type ControlUUID = string;
type ControlTime = number;
type ControlJSON = any;

declare namespace NodeJS {
  export interface Process {
    browser: boolean;
  }
}
