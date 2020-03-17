import fs from "fs";
import os from "os";

import { DEV } from "./config";

interface ClientOptions {
  secret?: string;
}

export default class Client {
  public secret: string | null;

  constructor(opts?: ClientOptions) {
    this.secret = opts && opts.secret || this.readCLISecret();
  }

  private readCLISecret() {
    const secretPath = `${os.homedir()}/.beneath/${DEV ? "secret_dev" : "secret"}.txt`;
    if (fs.existsSync(secretPath)) {
      const secret = fs.readFileSync(secretPath, "utf8");
      if (secret) {
        return secret;
      }
    }
    return null;
  }

}
