import { ApolloError } from "apollo-server";

export class ArgsError extends ApolloError {
  constructor(message: string) {
    super(message, "BAD_ARGS");
    Object.defineProperty(this, "name", { value: "ArgsError" });
  }
}
