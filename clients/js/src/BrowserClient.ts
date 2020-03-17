interface BrowserClientOptions {
  secret?: string;
}

export default class BrowserClient {
  public secret: string | undefined;

  constructor(opts?: BrowserClientOptions) {
    this.secret = opts && opts.secret;
  }

}
