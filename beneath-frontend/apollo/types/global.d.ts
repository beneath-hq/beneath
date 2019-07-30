// GraphQL (control) custom scalars
type ControlUUID = string;
type ControlTime = number;

declare namespace NodeJS {
  export interface Process {
    browser: boolean;
  }
}
