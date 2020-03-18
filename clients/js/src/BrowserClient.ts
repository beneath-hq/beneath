interface BrowserClientOptions {
  secret?: string;
}

export class BrowserClient {
  public secret: string | undefined;

  constructor(opts?: BrowserClientOptions) {
    this.secret = opts && opts.secret;
  }

}

/*

USE CASES

- figure out frontend needs
- write react hooks
- then write the client

LATER

- worry about server-side loads/writes (node.js)
- worry about creating one to X mapping models in Node

*/

