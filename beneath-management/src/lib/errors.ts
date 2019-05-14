import { ApolloError, UserInputError } from "apollo-server";
import { ValidationError as ClassValidatorValidationError } from "class-validator";


export class ArgsError extends ApolloError {
  constructor(message: string) {
    super(message, "BAD_ARGS");
    Object.defineProperty(this, "name", { value: "ArgsError" });
  }
}

export class NotFoundError extends ApolloError {
  constructor(message: string) {
    super(message, "NOT_FOUND");
    Object.defineProperty(this, "name", { value: "NotFoundError" });
  }
}

export class ValidationError extends ApolloError {
  constructor(message: string, errors: ClassValidatorValidationError[]) {
    const fields = errors.map((validationError) => validationError.property);
    super(message, "VALIDATION_ERROR", { fields });
    Object.defineProperty(this, "name", { value: "ValidationError" });
  }
}
