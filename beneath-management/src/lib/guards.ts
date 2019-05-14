import { validate } from "class-validator";
import { BaseEntity } from "typeorm";
import { ArgsError, ValidationError } from "../lib/errors";

export const areExclusiveArgs = (args: any, keys: string[]) => {
  const keysPresent = keys.map((key) => args[key] ? 1 : 0).reduce((a, b) => a + b, 0);
  return keysPresent === 1;
};

export const requireExclusiveArgs = (args: any, keys: string[]) => {
  if (!areExclusiveArgs(args, keys)) {
    throw new ArgsError(`Set one and only one of these args: ${JSON.stringify(keys)}`);
  }
};

export const requireValidates = async (object: BaseEntity) => {
  const errors = await validate(object);
  if (errors.length > 0) {
    throw new ValidationError("Validation error", errors);
  }
};
